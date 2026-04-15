package coingecko

type simplePriceResponse struct {
	Tether  map[string]float64 `json:"tether"`
	Bitcoin map[string]float64 `json:"bitcoin"`
}
