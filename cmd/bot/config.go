package main

import (
	"os"
	"strconv"
	"time"

	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/cbr"
)

const (
	defaultFallbackRate = 80
	defaultCBRCacheTTL  = time.Hour
)

// Config — все параметры запуска бота, собранные из переменных окружения.
// Заполняется один раз в loadConfig() и дальше read-only.
type Config struct {
	BotToken     string
	FallbackRate float64
	CBRCacheTTL  time.Duration
	CBRRateURL   string
}

// loadConfig читает конфигурацию из окружения.
// Невалидные значения молча заменяются на дефолтные — упасть на старте здесь
// смысла нет, бот всё равно работает с любым валидным fallback'ом.
func loadConfig() Config {
	return Config{
		BotToken:     os.Getenv("BOT_TOKEN"),
		FallbackRate: envFloat("USD_RUB_FALLBACK", defaultFallbackRate),
		CBRCacheTTL:  envDuration("CBR_CACHE_TTL", defaultCBRCacheTTL),
		CBRRateURL:   envString("CBR_RATE_URL", cbr.DefaultDailyURL),
	}
}

func envString(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

func envDuration(name string, fallback time.Duration) time.Duration {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}

func envFloat(name string, fallback float64) float64 {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
