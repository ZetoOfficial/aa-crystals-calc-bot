package main

import (
	"os"
	"time"

	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/coingecko"
)

const (
	defaultCoingeckoBaseURL = coingecko.DefaultBaseURL
	defaultRatesCacheTTL    = time.Hour
)

// Config — все параметры запуска бота, собранные из переменных окружения.
// Заполняется один раз в loadConfig() и дальше read-only.
type Config struct {
	BotToken         string
	CoingeckoBaseURL string
	RatesCacheTTL    time.Duration
}

// loadConfig читает конфигурацию из окружения.
// Невалидные значения молча заменяются на дефолтные — упасть на старте здесь
// смысла нет, бот всё равно работает с любым валидным fallback'ом.
func loadConfig() Config {
	return Config{
		BotToken:         os.Getenv("BOT_TOKEN"),
		CoingeckoBaseURL: envString("COINGECKO_BASE_URL", defaultCoingeckoBaseURL),
		RatesCacheTTL:    envDuration("RATES_CACHE_TTL", defaultRatesCacheTTL),
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
