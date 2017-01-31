package run

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type parseResult struct {
	PriceStr string
	DateStr  string
	Price    float32
	Date     time.Time
}

func parseDate(layout, str string) time.Time {
	loc, _ := time.LoadLocation("Europe/Rome")
	t, _ := time.ParseInLocation(layout, str, loc)
	return t
}

func parseFinanzaRepubblicaIt(doc *goquery.Document) (*parseResult, error) {
	var sPrice, sDate string

	doc.Find("div.TLB-scheda-body-container > ul > li:first-child > b ").Each(func(i int, s *goquery.Selection) {
		switch i {
		case 2:
			sPrice = s.Text()
		case 3:
			sDate = s.Text()
		}
	})
	if sPrice == "" {
		return nil, errors.New("Price not found")
	}

	price, err := strconv.ParseFloat(strings.Replace(sPrice, ",", ".", 1), 32)
	if err != nil {
		return nil, err
	}

	res := &parseResult{
		PriceStr: sPrice,
		DateStr:  sDate,
		Date:     parseDate("02/01/2006", sDate),
		Price:    float32(price),
	}

	return res, nil

}
