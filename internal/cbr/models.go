package cbr

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
)

// Rate — курс одной валюты к рублю на дату ЦБ.
type Rate struct {
	Code     string
	Name     string
	Nominal  int
	Value    float64 // RUB за Nominal единиц
	UnitRate float64 // RUB за 1 единицу
	Date     time.Time
}

// dailyResponse — корень XML-ответа /scripts/XML_daily.asp.
type dailyResponse struct {
	XMLName xml.Name    `xml:"ValCurs"`
	Date    string      `xml:"Date,attr"`
	Name    string      `xml:"name,attr"`
	Valutes []dailyItem `xml:"Valute"`
}

// dailyItem — один блок Valute в XML-ответе.
type dailyItem struct {
	ID       string `xml:"ID,attr"`
	NumCode  string `xml:"NumCode"`
	CharCode string `xml:"CharCode"`
	Nominal  string `xml:"Nominal"`
	Name     string `xml:"Name"`
	Value    string `xml:"Value"`
}

// parseDecimal парсит число вида "81,2500" (запятая как разделитель) в float64.
func parseDecimal(s string) (float64, error) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", "."))
	return strconv.ParseFloat(s, 64)
}

// parseCBRDate парсит дату ЦБ в одном из известных форматов.
func parseCBRDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range []string{"02.01.2006", "02/01/2006"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported date format: %q", s)
}

// charsetReader подключается к xml.Decoder для поддержки windows-1251.
// ЦБ отдаёт XML именно в этой кодировке.
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(strings.TrimSpace(charset)) {
	case "utf-8", "utf8":
		return input, nil
	case "windows-1251", "cp1251":
		return charmap.Windows1251.NewDecoder().Reader(input), nil
	default:
		return nil, fmt.Errorf("unsupported charset: %s", charset)
	}
}
