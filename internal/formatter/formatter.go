package formatter

import (
	"fmt"
	"strings"

	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/calculator"
)

const HelpText = "Отправь количество ШК.\nПример: 100"

type Service struct{}

func (Service) Result(result calculator.Result) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Для %d ШК нужно %d кристаллов.\n\n", result.SHK, result.TargetCrystals)
	b.WriteString("Выгоднее всего купить:\n")
	for _, choice := range result.Combo {
		fmt.Fprintf(&b, "- %d x %d USD = %d кристаллов\n", choice.Count, choice.Pack.USD, choice.Count*choice.Pack.Crystals)
	}

	fmt.Fprintf(&b, "\nИтого:\n")
	fmt.Fprintf(&b, "- %d кристаллов\n", result.TotalCrystals)
	fmt.Fprintf(&b, "- %d USDT\n", result.TotalUSDT)
	if result.RatesAvailable {
		fmt.Fprintf(&b, "- %.0f RUB\n", result.TotalRUB)
		fmt.Fprintf(&b, "- %.8f BTC\n\n", result.TotalBTC)
		fmt.Fprintf(&b, "- %.8f RUB per 1 ШК\n", result.RUBPerSHK)
		fmt.Fprintf(&b, "- %.8f USDT per 1 ШК\n", result.USDTPerSHK)
		fmt.Fprintf(&b, "- %.8f BTC per 1 ШК\n\n", result.BTCPerSHK)
	} else {
		b.WriteString("- RUB: курс временно недоступен\n")
		b.WriteString("- BTC: курс временно недоступен\n\n")
	}
	fmt.Fprintf(&b, "Излишек: %d кристаллов.", result.ExtraCrystals)

	return b.String()
}
