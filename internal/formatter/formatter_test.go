package formatter

import (
	"strings"
	"testing"

	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/calculator"
)

func TestResult(t *testing.T) {
	result := calculator.Result{
		SHK:            100,
		TargetCrystals: 105000,
		TotalCrystals:  105600,
		TotalUSD:       425,
		TotalRUB:       34000,
		ExtraCrystals:  600,
		Combo:          map[int]int{100: 4, 20: 1, 5: 1},
		Packs:          calculator.DefaultPacks,
	}

	got := Result(result)
	for _, want := range []string{
		"Для 100 ШК нужно 105000 кристаллов.",
		"- 4 x 100 USD = 100000 кристаллов",
		"- 1 x 20 USD = 4600 кристаллов",
		"- 425 USD",
		"- 34000 RUB",
		"Излишек: 600 кристаллов.",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("Result() missing %q in:\n%s", want, got)
		}
	}

	if strings.Contains(got, "50 USD") {
		t.Fatalf("Result() contains zero-count package:\n%s", got)
	}
}
