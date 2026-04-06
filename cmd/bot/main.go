package main

import (
	"context"
	_ "expvar"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/cbr"
	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/tgbot"
)

const (
	defaultFallbackRate = 80
	defaultMaxSHK       = 100000
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	fallbackRate := envFloat("USD_RUB_FALLBACK", defaultFallbackRate)
	maxSHK := envInt("MAX_SHK", defaultMaxSHK)
	metricsAddr := envString("METRICS_ADDR", "127.0.0.1:8080")

	rateProvider := &cbr.CachedClient{
		Client:   cbr.NewWithBaseURL(&http.Client{Timeout: 5 * time.Second}, envString("CBR_RATE_URL", cbr.DefaultDailyURL)),
		Fallback: fallbackRate,
	}

	token := os.Getenv("BOT_TOKEN")
	bot, err := tgbot.NewBot(token, tgbot.Handler{
		RateProvider: rateProvider,
		MaxSHK:       maxSHK,
		Logger:       logger,
	})
	if err != nil {
		logger.Printf("failed to create bot: %v", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	metricsServer := startMetricsServer(metricsAddr, logger)
	go func() {
		<-ctx.Done()
		logger.Print("stopping bot")
		if metricsServer != nil {
			if err := metricsServer.Shutdown(context.Background()); err != nil {
				logger.Printf("metrics server shutdown failed: %v", err)
			}
		}
		bot.Stop()
	}()

	logger.Print("bot started")
	bot.Start()
	logger.Print("bot stopped")
}

func startMetricsServer(addr string, logger *log.Logger) *http.Server {
	if addr == "" || addr == "off" || addr == "disabled" {
		logger.Print("metrics server disabled")
		return nil
	}

	server := &http.Server{Addr: addr}
	go func() {
		logger.Printf("metrics server started addr=%s endpoint=http://%s/debug/vars", addr, addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Printf("metrics server failed: %v", err)
		}
	}()
	return server
}

func envString(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

func envInt(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
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
