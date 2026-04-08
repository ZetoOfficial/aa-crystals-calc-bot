package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/calculator"
	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/cbr"
	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/formatter"
	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/parser"
	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/tgbot"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cfg := loadConfig()

	cbrClient := cbr.NewWithBaseURL(&http.Client{Timeout: 5 * time.Second}, cfg.CBRRateURL)
	rateProvider := cbr.NewCachingProvider(cbrClient, cfg.CBRCacheTTL, cfg.FallbackRate)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	handler := &tgbot.Handler{
		Rates:      rateProvider,
		Parser:     parser.Service{},
		Calculator: calculator.NewService(),
		Formatter:  formatter.Service{},
		RootCtx:    ctx,
	}
	bot, err := tgbot.NewBot(cfg.BotToken, handler)
	if err != nil {
		slog.Error("failed to create bot", "err", err)
		os.Exit(1)
	}

	go func() {
		<-ctx.Done()
		slog.Info("stopping bot")
		bot.Stop()
		handler.Wait()
	}()

	slog.Info("bot started")
	bot.Start()
	slog.Info("bot stopped")
}
