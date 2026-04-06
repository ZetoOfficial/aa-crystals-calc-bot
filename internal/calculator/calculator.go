package calculator

import (
	"errors"
	"fmt"
	"math"
)

const CrystalsPerSHK = 1050

var (
	ErrInvalidSHK = errors.New("shk must be greater than zero")
	ErrNoPacks    = errors.New("packs list is empty")
)

type Pack struct {
	USD      int
	Crystals int
}

type Result struct {
	SHK            int
	TargetCrystals int
	TotalCrystals  int
	TotalUSD       int
	TotalRUB       int
	ExtraCrystals  int
	Combo          map[int]int
	Packs          []Pack
}

var DefaultPacks = []Pack{
	{USD: 100, Crystals: 25000},
	{USD: 50, Crystals: 11800},
	{USD: 30, Crystals: 7000},
	{USD: 20, Crystals: 4600},
	{USD: 10, Crystals: 2200},
	{USD: 5, Crystals: 1000},
	{USD: 1, Crystals: 180},
}

func Calculate(shk int, usdRubRate float64) (Result, error) {
	return CalculateWithPacks(shk, usdRubRate, DefaultPacks)
}

func CalculateWithPacks(shk int, usdRubRate float64, packs []Pack) (Result, error) {
	if shk <= 0 {
		return Result{}, ErrInvalidSHK
	}
	if len(packs) == 0 {
		return Result{}, ErrNoPacks
	}

	target := shk * CrystalsPerSHK
	maxPack, err := bestEfficiencyPack(packs)
	if err != nil {
		return Result{}, err
	}

	maxCost := ceilDiv(target, maxPack.Crystals) * maxPack.USD
	maxByCost := buildMaxByCost(packs, maxCost)

	bestCost := -1
	for cost := 0; cost <= maxCost; cost++ {
		if maxByCost[0][cost] >= target {
			bestCost = cost
			break
		}
	}
	if bestCost < 0 {
		return Result{}, fmt.Errorf("no package combination covers %d crystals", target)
	}

	bestCrystals, combo := minimizeCrystalsForCost(packs, maxByCost, bestCost, target)
	totalRUB := int(math.Round(float64(bestCost) * usdRubRate))

	return Result{
		SHK:            shk,
		TargetCrystals: target,
		TotalCrystals:  bestCrystals,
		TotalUSD:       bestCost,
		TotalRUB:       totalRUB,
		ExtraCrystals:  bestCrystals - target,
		Combo:          combo,
		Packs:          clonePacks(packs),
	}, nil
}

func bestEfficiencyPack(packs []Pack) (Pack, error) {
	var best Pack
	for _, pack := range packs {
		if pack.USD <= 0 || pack.Crystals <= 0 {
			return Pack{}, fmt.Errorf("invalid pack: %+v", pack)
		}
		if best.USD == 0 || pack.Crystals*best.USD > best.Crystals*pack.USD {
			best = pack
		}
	}
	return best, nil
}

func buildMaxByCost(packs []Pack, maxCost int) [][]int {
	maxByCost := make([][]int, len(packs)+1)
	maxByCost[len(packs)] = make([]int, maxCost+1)
	for cost := 1; cost <= maxCost; cost++ {
		maxByCost[len(packs)][cost] = -1
	}

	for i := len(packs) - 1; i >= 0; i-- {
		maxByCost[i] = make([]int, maxCost+1)
		copy(maxByCost[i], maxByCost[i+1])

		pack := packs[i]
		for cost := pack.USD; cost <= maxCost; cost++ {
			prev := maxByCost[i][cost-pack.USD]
			if prev < 0 {
				continue
			}
			if candidate := prev + pack.Crystals; candidate > maxByCost[i][cost] {
				maxByCost[i][cost] = candidate
			}
		}
	}

	return maxByCost
}

func minimizeCrystalsForCost(packs []Pack, maxByCost [][]int, cost int, target int) (int, map[int]int) {
	bestCrystals := maxByCost[0][cost]
	bestCombo := make([]int, len(packs))
	combo := make([]int, len(packs))

	var save = func(total int) {
		if total >= target && total <= bestCrystals {
			bestCrystals = total
			copy(bestCombo, combo)
		}
	}

	var dfs func(index int, remainingCost int, crystals int)
	dfs = func(index int, remainingCost int, crystals int) {
		if remainingCost < 0 || index >= len(packs) {
			return
		}
		if maxByCost[index][remainingCost] < 0 {
			return
		}
		if crystals+maxByCost[index][remainingCost] < target {
			return
		}

		pack := packs[index]
		if index == len(packs)-1 {
			if remainingCost%pack.USD != 0 {
				return
			}
			combo[index] = remainingCost / pack.USD
			save(crystals + combo[index]*pack.Crystals)
			combo[index] = 0
			return
		}

		for count := 0; count <= remainingCost/pack.USD; count++ {
			nextCost := remainingCost - count*pack.USD
			nextCrystals := crystals + count*pack.Crystals
			if nextCrystals+maxByCost[index+1][nextCost] < target {
				continue
			}
			if nextCrystals >= bestCrystals {
				break
			}

			combo[index] = count
			dfs(index+1, nextCost, nextCrystals)
		}
		combo[index] = 0
	}

	dfs(0, cost, 0)

	result := make(map[int]int, len(packs))
	for i, count := range bestCombo {
		if count > 0 {
			result[packs[i].USD] = count
		}
	}

	return bestCrystals, result
}

func ceilDiv(a, b int) int {
	return (a + b - 1) / b
}

func clonePacks(packs []Pack) []Pack {
	cloned := make([]Pack, len(packs))
	copy(cloned, packs)
	return cloned
}
