package cbr

import (
	"context"
	"sync"
	"time"
)

// rateSource — внутренний интерфейс, которому обязан удовлетворять upstream
// у CachingProvider. Не экспортируется специально: потребители cbr должны
// определять свои собственные интерфейсы у себя (idiomatic Go consumer-side).
type rateSource interface {
	USDRUB(ctx context.Context) (float64, error)
}

// CachingProvider оборачивает rateSource (например *Client), кэшируя последний
// успешный курс. Не отравляет кэш fallback'ом: если upstream недоступен и нет
// валидного предыдущего значения, возвращает Fallback (если задан), но в кэш
// его не пишет — следующий вызов снова пойдет в upstream.
type CachingProvider struct {
	Upstream rateSource
	// Fallback — курс, возвращаемый если upstream недоступен и нет валидного кэша.
	// Не сохраняется в кэше.
	Fallback float64
	// TTL задает время жизни закэшированного курса.
	// Нулевое значение означает, что кэш не истекает.
	TTL time.Duration

	mu          sync.Mutex
	lastRate    float64
	lastFetched time.Time
}

// NewCachingProvider создаёт CachingProvider с указанным upstream и параметрами.
func NewCachingProvider(upstream rateSource, ttl time.Duration, fallback float64) *CachingProvider {
	return &CachingProvider{
		Upstream: upstream,
		TTL:      ttl,
		Fallback: fallback,
	}
}

func (c *CachingProvider) USDRUB(ctx context.Context) (float64, error) {
	if rate, ok := c.cached(); ok {
		return rate, nil
	}

	rate, err := c.Upstream.USDRUB(ctx)
	if err == nil && rate > 0 {
		c.store(rate)
		return rate, nil
	}

	// Upstream недоступен — кэш fallback'ом не отравляем.
	if stale, ok := c.stale(); ok {
		return stale, err
	}
	if c.Fallback > 0 {
		return c.Fallback, err
	}
	return 0, err
}

// cached возвращает закэшированное значение, если оно валидно и не истекло.
func (c *CachingProvider) cached() (float64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lastRate <= 0 || c.expiredLocked() {
		return 0, false
	}
	return c.lastRate, true
}

// stale возвращает закэшированное значение, игнорируя TTL.
// Используется как «лучше старое, чем ничего» при недоступном upstream.
func (c *CachingProvider) stale() (float64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lastRate <= 0 {
		return 0, false
	}
	return c.lastRate, true
}

func (c *CachingProvider) store(rate float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastRate = rate
	c.lastFetched = time.Now()
}

// expiredLocked возвращает true, если закэшированный курс устарел.
// Предполагается, что мьютекс уже захвачен вызывающей стороной.
func (c *CachingProvider) expiredLocked() bool {
	if c.TTL <= 0 {
		return false
	}
	return time.Since(c.lastFetched) >= c.TTL
}
