package tgbot

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"log"
	"time"

	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/calculator"
	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/cbr"
	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/formatter"
	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/parser"

	tele "gopkg.in/telebot.v4"
)

var (
	botRequestsTotal  = expvar.NewMap("aa_bot_requests_total")
	botErrorsTotal    = expvar.NewMap("aa_bot_errors_total")
	botDurationUs     = expvar.NewMap("aa_bot_duration_us_total")
	botLastDurationUs = expvar.NewMap("aa_bot_last_duration_us")
)

type Handler struct {
	RateProvider cbr.Provider
	MaxSHK       int
	Logger       *log.Logger
}

type requestTrace struct {
	parseUs     int64
	rateUs      int64
	calculateUs int64
	formatUs    int64
	sendUs      int64
	answerUs    int64
}

func Register(bot *tele.Bot, handler Handler) {
	bot.Handle("/start", handler.Help)
	bot.Handle("/help", handler.Help)
	bot.Handle(tele.OnText, handler.OnText)
	bot.Handle(tele.OnQuery, handler.OnQuery)
}

func (h Handler) Help(c tele.Context) error {
	return c.Send(formatter.HelpText)
}

func (h Handler) OnText(c tele.Context) error {
	start := time.Now()
	botRequestsTotal.Add("text", 1)

	text := c.Text()
	if text == "помощь" {
		err, sendUs := h.timedSend(c, formatter.HelpText)
		h.logTrace("text", start, requestTrace{sendUs: sendUs})
		return err
	}

	response, trace, err := h.calculate(c, text)
	if errors.Is(err, parser.ErrUnknownCommand) {
		botErrorsTotal.Add("unknown_command", 1)
		h.logTrace("text", start, trace)
		return nil
	}
	if err != nil {
		botErrorsTotal.Add("calculate", 1)
		sendErr, sendUs := h.timedSend(c, parser.UserMessage(err))
		trace.sendUs = sendUs
		h.logTrace("text", start, trace)
		return sendErr
	}

	err, trace.sendUs = h.timedSend(c, response)
	h.logTrace("text", start, trace)
	return err
}

func (h Handler) OnQuery(c tele.Context) error {
	start := time.Now()
	botRequestsTotal.Add("inline", 1)

	query := c.Query()
	if query == nil {
		h.logTrace("inline", start, requestTrace{})
		return nil
	}

	response, trace, err := h.calculate(c, query.Text)
	if err != nil {
		botErrorsTotal.Add("calculate", 1)
		answerErr, answerUs := h.timedAnswerArticle(c, "Формат: 100", parser.UserMessage(err), formatter.HelpText, "help")
		trace.answerUs = answerUs
		h.logTrace("inline", start, trace)
		return answerErr
	}

	err, trace.answerUs = h.timedAnswerArticle(c, "Расчет ШК", "Готовый расчет стоимости через пакеты", response, "calc")
	h.logTrace("inline", start, trace)
	return err
}

func (h Handler) calculate(c tele.Context, input string) (string, requestTrace, error) {
	var trace requestTrace

	parseStart := time.Now()
	command, err := parser.Parse(input, h.MaxSHK)
	trace.parseUs = recordBotDuration("parse", parseStart)
	if err != nil {
		return "", trace, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rateStart := time.Now()
	rate, rateErr := h.RateProvider.USDRUB(ctx)
	trace.rateUs = recordBotDuration("rate", rateStart)
	if rateErr != nil && h.Logger != nil {
		h.Logger.Printf("failed to fetch USD/RUB rate: %v", rateErr)
	}

	calculateStart := time.Now()
	result, err := calculator.Calculate(command.SHK, rate)
	trace.calculateUs = recordBotDuration("calculate", calculateStart)
	if err != nil {
		return "", trace, err
	}

	formatStart := time.Now()
	response := formatter.Result(result)
	trace.formatUs = recordBotDuration("format", formatStart)
	return response, trace, nil
}

func (h Handler) timedSend(c tele.Context, text string) (error, int64) {
	start := time.Now()
	err := c.Send(text)
	duration := recordBotDuration("send", start)
	if err != nil {
		botErrorsTotal.Add("send", 1)
	}
	return err, duration
}

func (h Handler) timedAnswerArticle(c tele.Context, title string, description string, text string, id string) (error, int64) {
	start := time.Now()
	err := h.answerArticle(c, title, description, text, id)
	duration := recordBotDuration("answer", start)
	if err != nil {
		botErrorsTotal.Add("answer", 1)
	}
	return err, duration
}

func (h Handler) logTrace(kind string, start time.Time, trace requestTrace) {
	total := recordBotDuration("total", start)
	if h.Logger != nil {
		h.Logger.Printf(
			"request_trace kind=%s total_ms=%.3f parse_ms=%.3f rate_ms=%.3f calculate_ms=%.3f format_ms=%.3f send_ms=%.3f answer_ms=%.3f",
			kind,
			usToMS(total),
			usToMS(trace.parseUs),
			usToMS(trace.rateUs),
			usToMS(trace.calculateUs),
			usToMS(trace.formatUs),
			usToMS(trace.sendUs),
			usToMS(trace.answerUs),
		)
	}
}

func (h Handler) answerArticle(c tele.Context, title string, description string, text string, id string) error {
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

func NewBot(token string, handler Handler) (*tele.Bot, error) {
	if token == "" {
		return nil, fmt.Errorf("BOT_TOKEN is required")
	}

	bot, err := tele.NewBot(tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
		OnError: func(err error, c tele.Context) {
			if handler.Logger != nil {
				handler.Logger.Printf("telegram handler error: %v", err)
			}
		},
	})
	if err != nil {
		return nil, err
	}

	Register(bot, handler)
	return bot, nil
}

func recordBotDuration(stage string, start time.Time) int64 {
	duration := time.Since(start).Microseconds()
	botDurationUs.Add(stage, duration)
	botLastDurationUs.Set(stage, expvar.Func(func() interface{} {
		return duration
	}))
	return duration
}

func usToMS(us int64) float64 {
	return float64(us) / 1000
}
