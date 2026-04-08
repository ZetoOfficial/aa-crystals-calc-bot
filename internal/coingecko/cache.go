package coingecko

import (
	"context"
	"sync"
	"time"

	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/calculator"
)

// rateSource — внутренний интерфейс, которому обязан удовлетворять upstream
// у CachingProvider. Не экспортируется специально: потребители coingecko
// должны определять свои собственные интерфейсы у себя (idiomatic Go consumer-side).
type rateSource interface {
	Rates(ctx context.Context) (calculator.Rates, error)
}

// CachingProvider оборачивает rateSource (например *Client), кэшируя
// последний успешный набор курсов. Если upstream упал и в кэше есть
// прошлое значение — отдаёт его (вместе с ошибкой), иначе возвращает
// нулевые курсы и ошибку. Никаких статических fallback'ов.
type CachingProvider struct {
	Upstream rateSource
	// TTL задаёт время жизни закэшированных курсов.
	// Нулевое значение означает, что кэш не истекает.
	TTL time.Duration

	mu          sync.Mutex
	lastRates   calculator.Rates
	lastFetched time.Time
	hasLast     bool
}

// NewCachingProvider создаёт CachingProvider с указанным upstream и TTL.
func NewCachingProvider(upstream rateSource, ttl time.Duration) *CachingProvider {
	return &CachingProvider{
		Upstream: upstream,
		TTL:      ttl,
	}
}

func (c *CachingProvider) Rates(ctx context.Context) (calculator.Rates, error) {
	if rates, ok := c.cached(); ok {
		return rates, nil
	}

	rates, err := c.Upstream.Rates(ctx)
	if err == nil && rates.USDTRUB > 0 && rates.BTCRUB > 0 {
		c.store(rates)
		return rates, nil
	}

	// Upstream недоступен — возвращаем последнее известное значение, если есть.
	if stale, ok := c.stale(); ok {
		return stale, err
	}
	return calculator.Rates{}, err
}

// cached возвращает закэшированные курсы, если они валидны и не истекли.
func (c *CachingProvider) cached() (calculator.Rates, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.hasLast || c.expiredLocked() {
		return calculator.Rates{}, false
	}
	return c.lastRates, true
}

// stale возвращает закэшированные курсы, игнорируя TTL.
// Используется как «лучше старое, чем ничего» при недоступном upstream.
func (c *CachingProvider) stale() (calculator.Rates, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.hasLast {
		return calculator.Rates{}, false
	}
	return c.lastRates, true
}

func (c *CachingProvider) store(rates calculator.Rates) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastRates = rates
	c.lastFetched = time.Now()
	c.hasLast = true
}

// expiredLocked возвращает true, если закэшированные курсы устарели.
// Предполагается, что мьютекс уже захвачен вызывающей стороной.
func (c *CachingProvider) expiredLocked() bool {
	if c.TTL <= 0 {
		return false
	}
	return time.Since(c.lastFetched) >= c.TTL
}
