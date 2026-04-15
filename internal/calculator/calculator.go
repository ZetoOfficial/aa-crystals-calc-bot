package calculator

import (
	"context"
	"fmt"
)

type Service struct {
	packs []Pack
}

func NewService() *Service {
	return &Service{packs: DefaultPacks}
}

// Calculate ищет минимальную по стоимости комбинацию пакетов,
// дающую не менее shk*CrystalsPerSHK кристаллов.
//
// Алгоритм: классический unbounded knapsack по стоимости.
// dp[c] = максимум кристаллов, достижимый ровно за c USD (или -1, если недостижимо).
// Параллельно ведём parent[c] — индекс пакета, которым пришли в это состояние,
// чтобы восстановить комбинацию обратным проходом. Останавливаемся на первом
// c, где dp[c] >= target — это и есть оптимальная стоимость.
func (s *Service) Calculate(ctx context.Context, shk int, rates Rates) (Result, error) {
	packs := s.packs
	if shk <= 0 {
		return Result{}, ErrInvalidSHK
	}
	if len(packs) == 0 {
		return Result{}, ErrNoPacks
	}
	for _, p := range packs {
		if p.USD <= 0 || p.Crystals <= 0 {
			return Result{}, fmt.Errorf("invalid pack: %+v", p)
		}
	}

	target := shk * CrystalsPerSHK

	// Верхняя граница стоимости: купить только лучшим по эффективности паком.
	bestPack := mostEfficientPack(packs)
	maxCost := ceilDiv(target, bestPack.Crystals) * bestPack.USD

	dp := make([]int, maxCost+1)
	parent := make([]int, maxCost+1)
	for c := 1; c <= maxCost; c++ {
		dp[c] = -1
		parent[c] = -1
	}

	const ctxCheckEvery = 4096
	bestCost := -1
	for c := 1; c <= maxCost; c++ {
		if c%ctxCheckEvery == 0 {
			if err := ctx.Err(); err != nil {
				return Result{}, err
			}
		}
		for i, pack := range packs {
			if c < pack.USD {
				continue
			}
			prev := dp[c-pack.USD]
			if prev < 0 {
				continue
			}
			if cand := prev + pack.Crystals; cand > dp[c] {
				dp[c] = cand
				parent[c] = i
			}
		}
		if dp[c] >= target {
			bestCost = c
			break
		}
	}

	if bestCost < 0 {
		return Result{}, fmt.Errorf("no package combination covers %d crystals", target)
	}

	counts := make([]int, len(packs))
	for c := bestCost; c > 0; {
		i := parent[c]
		if i < 0 {
			return Result{}, fmt.Errorf("reconstruct combination at cost %d", c)
		}
		counts[i]++
		c -= packs[i].USD
	}

	combo := make([]PackChoice, 0, len(packs))
	for i, count := range counts {
		if count > 0 {
			combo = append(combo, PackChoice{Pack: packs[i], Count: count})
		}
	}

	totalCrystals := dp[bestCost]
	ratesAvailable := rates.USDTRUB > 0 && rates.BTCRUB > 0
	var totalRUB, totalBTC float64
	if ratesAvailable {
		totalRUB = float64(bestCost) * rates.USDTRUB
		totalBTC = totalRUB / rates.BTCRUB
	}

	rubPerSHK := 0.0
	if ratesAvailable && shk > 0 {
		rubPerSHK = totalRUB / float64(shk)
	}
	usdtPerSHK := 0.0
	if shk > 0 {
		usdtPerSHK = float64(bestCost) / float64(shk)
	}
	btcPerSHK := 0.0
	if ratesAvailable && shk > 0 {
		btcPerSHK = totalBTC / float64(shk)
	}

	return Result{
		SHK:            shk,
		TargetCrystals: target,
		TotalCrystals:  totalCrystals,
		TotalUSDT:      bestCost,
		TotalRUB:       totalRUB,
		TotalBTC:       totalBTC,
		RUBPerSHK:      rubPerSHK,
		USDTPerSHK:     usdtPerSHK,
		BTCPerSHK:      btcPerSHK,
		RatesAvailable: ratesAvailable,
		ExtraCrystals:  totalCrystals - target,
		Combo:          combo,
	}, nil
}

// mostEfficientPack возвращает пакет с лучшим отношением crystals/USD.
// Все пакеты уже провалидированы вызывающей стороной.
func mostEfficientPack(packs []Pack) Pack {
	best := packs[0]
	for _, p := range packs[1:] {
		// p лучше, если p.Crystals/p.USD > best.Crystals/best.USD,
		// что равно p.Crystals*best.USD > best.Crystals*p.USD без деления.
		if p.Crystals*best.USD > best.Crystals*p.USD {
			best = p
		}
	}
	return best
}

func ceilDiv(a, b int) int {
	return (a + b - 1) / b
}
