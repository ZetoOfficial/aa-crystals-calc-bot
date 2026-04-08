package cbr

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const DefaultDailyURL = "https://www.cbr.ru/scripts/XML_daily.asp"

// Client — HTTP-клиент к scripts/XML_daily.asp ЦБ РФ.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

func New(httpClient *http.Client) *Client {
	return NewWithBaseURL(httpClient, DefaultDailyURL)
}

func NewWithBaseURL(httpClient *http.Client, baseURL string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultDailyURL
	}
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// GetRate возвращает курс валюты к рублю.
// Например, для USD вернет UnitRate = сколько RUB за 1 USD.
func (c *Client) GetRate(ctx context.Context, code string, date *time.Time) (Rate, error) {
	if c == nil {
		c = New(nil)
	}

	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return Rate{}, fmt.Errorf("empty currency code")
	}

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return Rate{}, fmt.Errorf("parse base url: %w", err)
	}

	q := u.Query()
	if date != nil {
		q.Set("date_req", date.Format("02/01/2006")) // dd/mm/yyyy
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Rate{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "aa-crystals-calc-bot/0.1")
	req.Header.Set("Accept", "application/xml,text/xml,*/*")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Rate{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return Rate{}, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var dr dailyResponse
	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charsetReader
	if err := decoder.Decode(&dr); err != nil {
		return Rate{}, fmt.Errorf("decode xml: %w", err)
	}

	docDate, err := parseCBRDate(dr.Date)
	if err != nil {
		return Rate{}, fmt.Errorf("parse response date: %w", err)
	}

	for _, v := range dr.Valutes {
		if strings.EqualFold(strings.TrimSpace(v.CharCode), code) {
			nominal, err := strconv.Atoi(strings.TrimSpace(v.Nominal))
			if err != nil {
				return Rate{}, fmt.Errorf("parse nominal for %s: %w", code, err)
			}

			value, err := parseDecimal(v.Value)
			if err != nil {
				return Rate{}, fmt.Errorf("parse value for %s: %w", code, err)
			}

			if nominal <= 0 {
				return Rate{}, fmt.Errorf("invalid nominal for %s: %d", code, nominal)
			}

			return Rate{
				Code:     strings.ToUpper(v.CharCode),
				Name:     strings.TrimSpace(v.Name),
				Nominal:  nominal,
				Value:    value,
				UnitRate: value / float64(nominal),
				Date:     docDate,
			}, nil
		}
	}

	return Rate{}, fmt.Errorf("currency %s not found", code)
}

func (c *Client) USDRUB(ctx context.Context) (float64, error) {
	r, err := c.GetRate(ctx, "USD", nil)
	if err != nil {
		return 0, err
	}
	return r.UnitRate, nil
}
