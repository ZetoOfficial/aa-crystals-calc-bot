package formatter

import (
	"fmt"
	"strings"

	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/calculator"
)

const HelpText = "Отправь количество ШК.\nПример: 100"

// Service — реализация форматирования ответов бота.
// Безсостоятельная: можно использовать через нулевое значение.
type Service struct{}

// Result рендерит результат расчёта в текст для отправки пользователю.
func (Service) Result(result calculator.Result) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Для %d ШК нужно %d кристаллов.\n\n", result.SHK, result.TargetCrystals)
	b.WriteString("Выгоднее всего купить:\n")
	for _, choice := range result.Combo {
		fmt.Fprintf(&b, "- %d x %d USD = %d кристаллов\n", choice.Count, choice.Pack.USD, choice.Count*choice.Pack.Crystals)
	}

	fmt.Fprintf(&b, "\nИтого:\n")
	fmt.Fprintf(&b, "- %d кристаллов\n", result.TotalCrystals)
	fmt.Fprintf(&b, "- %d USD\n", result.TotalUSD)
	if result.RateAvailable {
		fmt.Fprintf(&b, "- %d RUB\n\n", result.TotalRUB)
	} else {
		b.WriteString("- RUB: курс временно недоступен\n\n")
	}
	fmt.Fprintf(&b, "Излишек: %d кристаллов.", result.ExtraCrystals)

	return b.String()
}
