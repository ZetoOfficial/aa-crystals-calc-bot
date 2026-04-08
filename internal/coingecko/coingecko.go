package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/calculator"
)

// DefaultBaseURL — корень CoinGecko Public API v3.
const DefaultBaseURL = "https://api.coingecko.com/api/v3"

// Client — HTTP-клиент к CoinGecko /simple/price.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

func New(httpClient *http.Client) *Client {
	return NewWithBaseURL(httpClient, DefaultBaseURL)
}

func NewWithBaseURL(httpClient *http.Client, baseURL string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		httpClient: httpClient,
		baseURL:    strings.TrimRight(baseURL, "/"),
	}
}

// Rates делает один запрос к /simple/price?ids=tether,bitcoin&vs_currencies=rub
// и возвращает оба курса в одной структуре.
func (c *Client) Rates(ctx context.Context) (calculator.Rates, error) {
	if c == nil {
		c = New(nil)
	}

	u, err := url.Parse(c.baseURL + "/simple/price")
	if err != nil {
		return calculator.Rates{}, fmt.Errorf("parse base url: %w", err)
	}
	q := u.Query()
	q.Set("ids", "tether,bitcoin")
	q.Set("vs_currencies", "rub")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return calculator.Rates{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "aa-crystals-calc-bot/0.1")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return calculator.Rates{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return calculator.Rates{}, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var sp simplePriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&sp); err != nil {
		return calculator.Rates{}, fmt.Errorf("decode json: %w", err)
	}

	usdtRub, ok := sp.Tether["rub"]
	if !ok || usdtRub <= 0 {
		return calculator.Rates{}, fmt.Errorf("missing or invalid tether/rub in response")
	}
	btcRub, ok := sp.Bitcoin["rub"]
	if !ok || btcRub <= 0 {
		return calculator.Rates{}, fmt.Errorf("missing or invalid bitcoin/rub in response")
	}

	return calculator.Rates{
		USDTRUB: usdtRub,
		BTCRUB:  btcRub,
	}, nil
}
