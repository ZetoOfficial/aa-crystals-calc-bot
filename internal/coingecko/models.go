package coingecko

// simplePriceResponse — ответ /api/v3/simple/price.
// Пример при ids=tether,bitcoin&vs_currencies=rub:
//
//	{"tether":{"rub":95.12},"bitcoin":{"rub":6500000.0}}
type simplePriceResponse struct {
	Tether  map[string]float64 `json:"tether"`
	Bitcoin map[string]float64 `json:"bitcoin"`
}
