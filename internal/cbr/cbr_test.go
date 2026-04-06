package cbr

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientGetUSDRUB(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.UserAgent() == "" {
			t.Fatal("missing user-agent")
		}
		fmt.Fprint(w, `<?xml version="1.0" encoding="windows-1251"?>
<ValCurs Date="06.04.2026" name="Foreign Currency Market">
	<Valute ID="R01235">
		<NumCode>840</NumCode>
		<CharCode>USD</CharCode>
		<Nominal>1</Nominal>
		<Name>Доллар США</Name>
		<Value>81,2500</Value>
	</Valute>
</ValCurs>`)
	}))
	defer server.Close()

	client := NewWithBaseURL(server.Client(), server.URL)
	rate, err := client.GetUSDRUB(context.Background())
	if err != nil {
		t.Fatalf("GetUSDRUB() error = %v", err)
	}
	if rate != 81.25 {
		t.Fatalf("rate = %v, want 81.25", rate)
	}
}

func TestCachedClientCachesSuccessfulRate(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		fmt.Fprintf(w, `<?xml version="1.0" encoding="windows-1251"?>
<ValCurs Date="06.04.2026" name="Foreign Currency Market">
	<Valute ID="R01235">
		<NumCode>840</NumCode>
		<CharCode>USD</CharCode>
		<Nominal>1</Nominal>
		<Name>Доллар США</Name>
		<Value>%d,0000</Value>
	</Valute>
</ValCurs>`, 80+calls)
	}))
	defer server.Close()

	provider := &CachedClient{
		Client: NewWithBaseURL(server.Client(), server.URL),
	}

	first, err := provider.USDRUB(context.Background())
	if err != nil {
		t.Fatalf("USDRUB() error = %v", err)
	}
	second, err := provider.USDRUB(context.Background())
	if err != nil {
		t.Fatalf("USDRUB() error = %v", err)
	}

	if first != 81 || second != 81 {
		t.Fatalf("rates = %v, %v; want 81, 81", first, second)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestCachedClientUsesFallback(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, "down", http.StatusBadGateway)
	}))
	defer server.Close()

	provider := &CachedClient{
		Client:   NewWithBaseURL(server.Client(), server.URL),
		Fallback: 80,
	}

	rate, err := provider.USDRUB(context.Background())
	if err == nil {
		t.Fatalf("USDRUB() error = nil, want upstream error")
	}
	if rate != 80 {
		t.Fatalf("rate = %v, want 80", rate)
	}

	rate, err = provider.USDRUB(context.Background())
	if err != nil {
		t.Fatalf("cached USDRUB() error = %v", err)
	}
	if rate != 80 {
		t.Fatalf("cached rate = %v, want 80", rate)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}
