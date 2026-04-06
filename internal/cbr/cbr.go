package cbr

import (
	"context"
	"encoding/xml"
	"expvar"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const DefaultDailyURL = "https://www.cbr.ru/scripts/XML_daily.asp"

var (
	cbrRequestsTotal  = expvar.NewMap("aa_cbr_requests_total")
	cbrDurationUs     = expvar.NewMap("aa_cbr_duration_us_total")
	cbrLastDurationUs = expvar.NewMap("aa_cbr_last_duration_us")
)

type Provider interface {
	USDRUB(ctx context.Context) (float64, error)
}

type Client struct {
	httpClient *http.Client
	baseURL    string
}

type CachedClient struct {
	Client   *Client
	Fallback float64

	mu       sync.Mutex
	lastRate float64
	hasLast  bool
}

type Rate struct {
	Code     string
	Name     string
	Nominal  int
	Value    float64 // RUB за Nominal единиц
	UnitRate float64 // RUB за 1 единицу
	Date     time.Time
}

type dailyResponse struct {
	XMLName xml.Name    `xml:"ValCurs"`
	Date    string      `xml:"Date,attr"`
	Name    string      `xml:"name,attr"`
	Valutes []dailyItem `xml:"Valute"`
}

type dailyItem struct {
	ID       string `xml:"ID,attr"`
	NumCode  string `xml:"NumCode"`
	CharCode string `xml:"CharCode"`
	Nominal  string `xml:"Nominal"`
	Name     string `xml:"Name"`
	Value    string `xml:"Value"`
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
	start := time.Now()
	defer recordCBRDuration("get_rate", start)

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

	httpStart := time.Now()
	resp, err := c.httpClient.Do(req)
	recordCBRDuration("http", httpStart)
	if err != nil {
		cbrRequestsTotal.Add("api_error", 1)
		return Rate{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		cbrRequestsTotal.Add("api_error", 1)
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

			cbrRequestsTotal.Add("api_success", 1)
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

func (c *Client) GetUSDRUB(ctx context.Context) (float64, error) {
	r, err := c.GetRate(ctx, "USD", nil)
	if err != nil {
		return 0, err
	}
	return r.UnitRate, nil
}

func (c *Client) USDRUB(ctx context.Context) (float64, error) {
	return c.GetUSDRUB(ctx)
}

func (c *CachedClient) USDRUB(ctx context.Context) (float64, error) {
	start := time.Now()
	defer recordCBRDuration("usdrub", start)

	c.mu.Lock()
	if c.hasLast {
		rate := c.lastRate
		c.mu.Unlock()
		cbrRequestsTotal.Add("cache_hit", 1)
		return rate, nil
	}
	c.mu.Unlock()

	cbrRequestsTotal.Add("cache_miss", 1)
	client := c.Client
	if client == nil {
		client = New(nil)
	}

	rate, err := client.GetUSDRUB(ctx)
	if err == nil && rate > 0 {
		c.mu.Lock()
		c.lastRate = rate
		c.hasLast = true
		c.mu.Unlock()
		cbrRequestsTotal.Add("cache_store_api", 1)
		return rate, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.hasLast && c.lastRate > 0 {
		return c.lastRate, err
	}
	if c.Fallback > 0 {
		c.lastRate = c.Fallback
		c.hasLast = true
		cbrRequestsTotal.Add("cache_store_fallback", 1)
		return c.Fallback, err
	}
	return 0, err
}

func recordCBRDuration(stage string, start time.Time) {
	duration := time.Since(start).Microseconds()
	cbrDurationUs.Add(stage, duration)
	cbrLastDurationUs.Set(stage, expvar.Func(func() interface{} {
		return duration
	}))
}

func parseDecimal(s string) (float64, error) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", "."))
	return strconv.ParseFloat(s, 64)
}

func parseCBRDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range []string{"02.01.2006", "02/01/2006"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported date format: %q", s)
}

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(strings.TrimSpace(charset)) {
	case "utf-8", "utf8":
		return input, nil
	case "windows-1251", "cp1251":
		data, err := io.ReadAll(input)
		if err != nil {
			return nil, err
		}
		return strings.NewReader(decodeWindows1251(data)), nil
	default:
		return nil, fmt.Errorf("unsupported charset: %s", charset)
	}
}

func decodeWindows1251(data []byte) string {
	var b strings.Builder
	b.Grow(len(data))
	for _, c := range data {
		switch {
		case c < 0x80:
			b.WriteByte(c)
		case c >= 0xC0:
			b.WriteRune(rune(c) + 0x0350)
		default:
			b.WriteRune(windows1251Table[c-0x80])
		}
	}
	return b.String()
}

var windows1251Table = [...]rune{
	0x0402, 0x0403, 0x201A, 0x0453, 0x201E, 0x2026, 0x2020, 0x2021,
	0x20AC, 0x2030, 0x0409, 0x2039, 0x040A, 0x040C, 0x040B, 0x040F,
	0x0452, 0x2018, 0x2019, 0x201C, 0x201D, 0x2022, 0x2013, 0x2014,
	0xFFFD, 0x2122, 0x0459, 0x203A, 0x045A, 0x045C, 0x045B, 0x045F,
	0x00A0, 0x040E, 0x045E, 0x0408, 0x00A4, 0x0490, 0x00A6, 0x00A7,
	0x0401, 0x00A9, 0x0404, 0x00AB, 0x00AC, 0x00AD, 0x00AE, 0x0407,
	0x00B0, 0x00B1, 0x0406, 0x0456, 0x0491, 0x00B5, 0x00B6, 0x00B7,
	0x0451, 0x2116, 0x0454, 0x00BB, 0x0458, 0x0405, 0x0455, 0x0457,
}
