package run

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type (
	// Scraper enum type
	ScraperType int

	scraperResult struct {
		price float32
		date  time.Time
	}

	parseDocFunc func(doc *goquery.Document) (*scraperResult, error)
)

// Scraper constants

const (
	Unknown ScraperType = iota
	BorsaItaliana
	FinanzaRepubblica
	MilanoFinanza
	SoldiOnline
)

var (
	s2i = map[string]ScraperType{
		"borsaitaliana":      BorsaItaliana,
		"finanza.repubblica": FinanzaRepubblica,
		"milanofinanza":      MilanoFinanza,
		"soldionline":        SoldiOnline,
	}
	i2s map[ScraperType]string
)

func init() {
	i2s = map[ScraperType]string{}
	for name, scraper := range s2i {
		i2s[scraper] = name
	}
	i2s[Unknown] = "<unknown>"
}

func ScraperTypeFromString(name string) (ScraperType, error) {
	var err error
	res := s2i[name]
	if res == Unknown {

		err = fmt.Errorf("Invalid scraper: %q", name)
	}
	return res, err
}

func ScraperTypeFromUrl(url string) (ScraperType, error) {
	var cases = []struct {
		prefix  string
		scrtype ScraperType
	}{
		{"http://www.borsaitaliana.it/", BorsaItaliana},
		{"http://finanza.repubblica.it/", FinanzaRepubblica},
		{"http://www.milanofinanza.it/", MilanoFinanza},
		{"http://www.soldionline.it/", SoldiOnline},
	}
	for _, item := range cases {
		if strings.HasPrefix(url, item.prefix) {
			return item.scrtype, nil
		}
	}
	return Unknown, fmt.Errorf("Can't get scraper from URL: %q", url)
}

func ScraperTypeFromStringOrUrl(name, url string) (ScraperType, error) {
	if name == "" {
		return ScraperTypeFromUrl(url)
	}
	return ScraperTypeFromString(name)
}

func (scraper ScraperType) String() string {
	if name, ok := i2s[scraper]; ok {
		return name
	}
	return fmt.Sprintf("ScraperType-<%d>", int(scraper))
}

func (scraper ScraperType) getParseDocFunc() parseDocFunc {
	return dummyParseDocFunc
}

func dummyParseDocFunc(doc *goquery.Document) (*scraperResult, error) {
	return nil, nil
}
