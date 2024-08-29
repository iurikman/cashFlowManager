package converter

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/iurikman/cashFlowManager/internal/models"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const cbrURL = "http://www.cbr.ru/scripts/XML_dynamic.asp?date_req1="

type Converter struct {
	currencyFrom Currency
	currencyTo   Currency
}

type ValCurs struct {
	ID         string   `xml:"ID,attr"`
	DateRange1 string   `xml:"DateRange1,attr"`
	DateRange2 string   `xml:"DateRange2,attr"`
	Name       string   `xml:"name,attr"`
	Records    []Record `xml:"Record"`
}

type Record struct {
	Date      string `xml:"Date,attr"`
	ID        string `xml:"Id,attr"`
	Nominal   int    `xml:"Nominal"`
	Value     string `xml:"Value"`
	VunitRate string `xml:"VunitRate"`
}

type Currency struct {
	Amount float64
	Name   string
}

func NewConverter(currencyFrom, currencyTo Currency) *Converter {
	return &Converter{
		currencyFrom: currencyFrom,
		currencyTo:   currencyTo,
	}
}

func (c *Converter) Convert(ctx context.Context) (float64, error) {
	codeCurrFrom := models.AllowedCurrencies[c.currencyFrom.Name]
	codeCurrTo := models.AllowedCurrencies[c.currencyTo.Name]

	switch {
	case c.currencyTo.Name == "RUR":
		changeRateCurrFrom, err := c.fetchRate(ctx, codeCurrFrom)
		if err != nil {
			return 0, fmt.Errorf("c.fetchRate(codeCurrFrom) err: %w", err)
		}

		result := c.currencyFrom.Amount * changeRateCurrFrom

		return result, nil
	default:
		changeRateCurrFrom, err := c.fetchRate(ctx, codeCurrFrom)
		if err != nil {
			return 0, fmt.Errorf("c.fetchRate(codeCurrFrom) err: %w", err)
		}

		changeRateCurrTo, err := c.fetchRate(ctx, codeCurrTo)
		if err != nil {
			return 0, fmt.Errorf("c.fetchRate(codeCurrTo) err: %w", err)
		}

		result := (c.currencyFrom.Amount * changeRateCurrFrom) / changeRateCurrTo

		return result, nil
	}
}

func (c *Converter) fetchRate(ctx context.Context, currencyCode string) (float64, error) {
	date := time.Now().AddDate(0, 0, -2)
	dateString := fmt.Sprintf("%02d/%02d/%d", date.Day(), date.Month(), date.Year())

	reqURLString := cbrURL + dateString + "&date_req2=" + dateString + "&VAL_NM_RQ=" + currencyCode

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURLString, nil)
	if err != nil {
		return 0, fmt.Errorf("http.NewRequest(\"GET\", reqURLString, nil) err: %w", err)
	}

	req.Header.Set("User-Agent", "YourAppName/1.0")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("http.Get(reqURLString) err: %w", err)
	}

	defer resp.Body.Close()

	reader := transform.NewReader(resp.Body, charmap.Windows1251.NewDecoder())

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = func(encoding string, input io.Reader) (io.Reader, error) {
		if encoding != "windows-1251" {
			return nil, fmt.Errorf("unsupported encoding: %s", encoding)
		}

		return transform.NewReader(input, charmap.Windows1251.NewDecoder()), nil
	}

	var valCurs ValCurs
	if err := decoder.Decode(&valCurs); err != nil {
		return 0, fmt.Errorf("xml.Unmarshal(err): %w", err)
	}

	if len(valCurs.Records) == 0 {
		return 0, fmt.Errorf("len(valCurs.Records) == 0 (currencyCode was %s)", currencyCode)
	}

	value := strings.ReplaceAll(valCurs.Records[0].Value, ",", ".")

	rate, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("strconv.ParseFloat(valCurs.Records[0].Value, 64) err: %w", err)
	}

	return rate, nil
}
