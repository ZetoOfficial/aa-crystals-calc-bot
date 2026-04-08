// Package tgbot реализует адаптер бота над telebot.v4.
package tgbot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/formatter"
	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/parser"

	tele "gopkg.in/telebot.v4"
)

// errHelpRequested — внутренний sentinel: пользователь попросил помощь.
// Не считается ошибкой расчёта, но удобно прокинуть как обычный возврат
// из calculate, чтобы хендлеры обработали единообразно.
var errHelpRequested = errors.New("help requested")

type Handler struct {
	Rates      RateProvider
	Parser     Parser
	Calculator Calculator
	Formatter  Formatter
	// RootCtx — корневой контекст бота. Все обработчики наследуют от него таймауты,
	// чтобы при остановке бота in-flight запросы корректно отменялись.
	RootCtx context.Context

	inflight sync.WaitGroup
}

// Wait блокируется, пока все in-flight обработчики не завершатся.
func (h *Handler) Wait() {
	h.inflight.Wait()
}

// track оборачивает обработчик: считает in-flight для graceful shutdown.
func (h *Handler) track(fn func(c tele.Context) error) func(tele.Context) error {
	return func(c tele.Context) error {
		h.inflight.Add(1)
		defer h.inflight.Done()
		return fn(c)
	}
}

func Register(bot *tele.Bot, handler *Handler) {
	bot.Handle("/start", handler.track(handler.Help))
	bot.Handle("/help", handler.track(handler.Help))
	bot.Handle(tele.OnText, handler.track(handler.OnText))
	bot.Handle(tele.OnQuery, handler.track(handler.OnQuery))
}

func (h *Handler) Help(c tele.Context) error {
	return c.Send(formatter.HelpText)
}

func (h *Handler) OnText(c tele.Context) error {
	response, err := h.calculate(c.Text())
	if errors.Is(err, errHelpRequested) {
		return c.Send(formatter.HelpText)
	}
	if err != nil {
		return c.Send(parser.UserMessage(err))
	}
	return c.Send(response)
}

func (h *Handler) OnQuery(c tele.Context) error {
	query := c.Query()
	if query == nil {
		return nil
	}

	response, err := h.calculate(query.Text)
	if errors.Is(err, errHelpRequested) {
		return h.answerArticle(c, "Помощь", "Как пользоваться ботом", formatter.HelpText, "help")
	}
	if err != nil {
		return h.answerArticle(c, "Формат: 100", parser.UserMessage(err), formatter.HelpText, "help")
	}
	return h.answerArticle(c, "Расчет ШК", "Готовый расчет стоимости через пакеты", response, "calc")
}

func (h *Handler) calculate(input string) (string, error) {
	command, err := h.Parser.Parse(input)
	if err != nil {
		return "", err
	}
	if command.Name == parser.CommandHelp {
		return "", errHelpRequested
	}

	root := h.RootCtx
	if root == nil {
		root = context.Background()
	}
	ctx, cancel := context.WithTimeout(root, 5*time.Second)
	defer cancel()

	rate, rateErr := h.Rates.USDRUB(ctx)
	if rateErr != nil {
		slog.Warn("failed to fetch USD/RUB rate", "err", rateErr)
	}

	result, err := h.Calculator.Calculate(ctx, command.SHK, rate)
	if err != nil {
		return "", err
	}

	return h.Formatter.Result(result), nil
}

func (h *Handler) answerArticle(c tele.Context, title string, description string, text string, id string) error {
	result := &tele.ArticleResult{
		Title:       title,
		Description: description,
		Text:        text,
	}
	result.SetResultID(id)

	return c.Answer(&tele.QueryResponse{
		Results:    tele.Results{result},
		CacheTime:  0,
		IsPersonal: true,
	})
}

func NewBot(token string, handler *Handler) (*tele.Bot, error) {
	if token == "" {
		return nil, fmt.Errorf("BOT_TOKEN is required")
	}

	bot, err := tele.NewBot(tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
		OnError: func(err error, c tele.Context) {
			slog.Error("telegram handler error", "err", err)
		},
	})
	if err != nil {
		return nil, err
	}

	Register(bot, handler)
	return bot, nil
}
