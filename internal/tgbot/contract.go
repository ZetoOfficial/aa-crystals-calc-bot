// Контракты, которые tgbot.Handler ожидает от внешнего мира.
// Все интерфейсы — consumer-side: реализуются другими пакетами или mock'ами.
//
//go:generate mockgen -source=contract.go -destination=mocks/mocks.go -package=mocks
package tgbot

import (
	"context"

	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/calculator"
	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/parser"
)

// RateProvider отдаёт текущий курс USD/RUB.
// Реализуется *cbr.Client и *cbr.CachingProvider, либо моком в тестах.
type RateProvider interface {
	USDRUB(ctx context.Context) (float64, error)
}

// Parser разбирает входное сообщение пользователя в команду.
type Parser interface {
	Parse(input string) (parser.Command, error)
}

// Calculator считает оптимальную комбинацию пакетов под заданное число ШК и курс.
type Calculator interface {
	Calculate(ctx context.Context, shk int, usdRubRate float64) (calculator.Result, error)
}

// Formatter рендерит результат расчёта в текст для пользователя.
type Formatter interface {
	Result(r calculator.Result) string
}
