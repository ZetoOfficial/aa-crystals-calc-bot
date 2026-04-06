package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ZetoOfficial/aa-crystals-calc-bot/internal/cbr"
)

func main() {
	ctx := context.Background()
	client := cbr.New(nil)

	usd, err := client.GetUSDRUB(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("USD/RUB: %.4f\n", usd)

	d := time.Now()
	eur, err := client.GetRate(ctx, "EUR", &d)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s on %s: %.4f RUB\n", eur.Code, eur.Date.Format("2006-01-02"), eur.UnitRate)
}
