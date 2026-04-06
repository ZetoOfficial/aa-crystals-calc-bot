package calculator

import "testing"

func TestCalculateFindsOptimalCombination(t *testing.T) {
	result, err := Calculate(100, 80)
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	if result.TargetCrystals != 105000 {
		t.Fatalf("TargetCrystals = %d, want 105000", result.TargetCrystals)
	}
	if result.TotalCrystals != 105140 {
		t.Fatalf("TotalCrystals = %d, want 105140, combo = %#v", result.TotalCrystals, result.Combo)
	}
	if result.TotalUSD != 423 {
		t.Fatalf("TotalUSD = %d, want 423", result.TotalUSD)
	}
	if result.TotalRUB != 33840 {
		t.Fatalf("TotalRUB = %d, want 33840", result.TotalRUB)
	}
	if result.ExtraCrystals != 140 {
		t.Fatalf("ExtraCrystals = %d, want 140", result.ExtraCrystals)
	}

	wantCombo := map[int]int{100: 4, 20: 1, 1: 3}
	for usd, want := range wantCombo {
		if got := result.Combo[usd]; got != want {
			t.Fatalf("Combo[%d] = %d, want %d", usd, got, want)
		}
	}
}

func TestCalculateMinimizesExtraAtSameUSD(t *testing.T) {
	result, err := CalculateWithPacks(1, 80, []Pack{
		{USD: 10, Crystals: 2000},
		{USD: 5, Crystals: 1050},
		{USD: 1, Crystals: 1},
	})
	if err != nil {
		t.Fatalf("CalculateWithPacks() error = %v", err)
	}

	if result.TotalUSD != 5 {
		t.Fatalf("TotalUSD = %d, want 5", result.TotalUSD)
	}
	if result.TotalCrystals != 1050 {
		t.Fatalf("TotalCrystals = %d, want 1050", result.TotalCrystals)
	}
	if result.ExtraCrystals != 0 {
		t.Fatalf("ExtraCrystals = %d, want 0", result.ExtraCrystals)
	}
}

func TestCalculateRejectsInvalidSHK(t *testing.T) {
	if _, err := Calculate(0, 80); err != ErrInvalidSHK {
		t.Fatalf("Calculate() error = %v, want %v", err, ErrInvalidSHK)
	}
}

func TestCalculateLargeInput(t *testing.T) {
	result, err := Calculate(100000, 80)
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}
	if result.TargetCrystals != 105000000 {
		t.Fatalf("TargetCrystals = %d, want 105000000", result.TargetCrystals)
	}
	if result.TotalUSD != 420000 {
		t.Fatalf("TotalUSD = %d, want 420000", result.TotalUSD)
	}
	if result.ExtraCrystals != 0 {
		t.Fatalf("ExtraCrystals = %d, want 0", result.ExtraCrystals)
	}
}
