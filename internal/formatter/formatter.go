package formatter

import (
	"fmt"
	"strings"

	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/calculator"
)

const HelpText = "Отправь количество ШК.\nПример: 100"

func Result(result calculator.Result) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Для %d ШК нужно %d кристаллов.\n\n", result.SHK, result.TargetCrystals)
	b.WriteString("Выгоднее всего купить:\n")
	for _, pack := range result.Packs {
		count := result.Combo[pack.USD]
		if count == 0 {
			continue
		}
		fmt.Fprintf(&b, "- %d x %d USD = %d кристаллов\n", count, pack.USD, count*pack.Crystals)
	}

	fmt.Fprintf(&b, "\nИтого:\n")
	fmt.Fprintf(&b, "- %d кристаллов\n", result.TotalCrystals)
	fmt.Fprintf(&b, "- %d USD\n", result.TotalUSD)
	fmt.Fprintf(&b, "- %d RUB\n\n", result.TotalRUB)
	fmt.Fprintf(&b, "Излишек: %d кристаллов.", result.ExtraCrystals)

	return b.String()
}
