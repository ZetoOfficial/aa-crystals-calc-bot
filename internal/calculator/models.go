package calculator

import "errors"

// CrystalsPerSHK — сколько кристаллов даёт одна ШК.
const CrystalsPerSHK = 1050

var (
	ErrInvalidSHK = errors.New("shk must be greater than zero")
	ErrNoPacks    = errors.New("packs list is empty")
)

// Pack — донат-пакет: цена в USD и сколько кристаллов даёт.
type Pack struct {
	USD      int
	Crystals int
}

// PackChoice — пакет и сколько его взяли в оптимальной комбинации.
type PackChoice struct {
	Pack  Pack
	Count int
}

// Rates — курсы для конвертации цены пакетов в RUB и BTC.
// Заполняется поставщиком курсов (например, coingecko.Client).
type Rates struct {
	USDTRUB float64 // RUB за 1 USDT
	BTCRUB  float64 // RUB за 1 BTC
}

// Result — итог расчёта оптимальной комбинации.
type Result struct {
	SHK            int
	TargetCrystals int
	TotalCrystals  int
	// TotalUSDT — сумма Pack.USD напрямую (1 USD = 1 USDT по допущению).
	TotalUSDT int
	// TotalRUB = TotalUSDT * Rates.USDTRUB. 0, если !RatesAvailable.
	TotalRUB float64
	// TotalBTC = TotalRUB / Rates.BTCRUB. 0, если !RatesAvailable.
	TotalBTC float64
	// RatesAvailable=false означает, что курсы получить не удалось
	// и TotalRUB/TotalBTC не следует показывать пользователю.
	RatesAvailable bool
	ExtraCrystals  int
	Combo          []PackChoice
}

// DefaultPacks — стандартный набор донат-пакетов игры.
var DefaultPacks = []Pack{
	{USD: 100, Crystals: 25000},
	{USD: 50, Crystals: 11800},
	{USD: 30, Crystals: 7000},
	{USD: 20, Crystals: 4600},
	{USD: 10, Crystals: 2200},
	// {USD: 5, Crystals: 1000}, карточек на 5 USD нет
	// {USD: 1, Crystals: 180}, карточек на 1 USD нет
}
