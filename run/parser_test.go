package run

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// getpath returns the relative path to the file "data-crypt/doc/<fname>".
// first it assumes the current working directory is the "getstock" folder.
// if the file does not exists, assumes the working directory is the "getstock/run" folder.
func getpath(fname string) string {
	const prefix = "data-crypt/doc"
	p := filepath.Join(prefix, fname)
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return filepath.Join("..", prefix, fname)
}

func newDocumentFromFile(path string) (*goquery.Document, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return goquery.NewDocumentFromReader(f)
}

func TestParsers(t *testing.T) {

	testCases := []struct {
		scraper  string
		priceStr string
		dateStr  string
	}{
		{"finanza.repubblica.it", "90,680", "22/12/2016"},
		{"www.borse.it", "5,3600", "20/01/2017"},
		{"www.eurotlx.com", "90,68", "30-01-2017"},
		{"www.milanofinanza.it", "5,048", "20/01/17 1.00.00"},
		{"www.morningstar.it", "5,158", "27/01/2017"},
		{"www.teleborsa.it", "5,368", "27/01/2017"},
	}

	for _, tc := range testCases {

		path := getpath(tc.scraper + ".html")
		doc, err := newDocumentFromFile(path)
		if err != nil {
			t.Error(tc.scraper, err)
			continue
		}
		parseFunc := getParseDocFunc(tc.scraper)
		res, err := parseFunc(doc)
		if err != nil {
			t.Error(tc.scraper, err)
			continue
		}
		t.Log(tc.scraper, "->", res)

		if res.PriceStr != tc.priceStr {
			t.Errorf("[%s] PriceStr: expected %q, found %q", tc.scraper, tc.priceStr, res.PriceStr)
		}
		if res.DateStr != tc.dateStr {
			t.Errorf("[%s] DateStr: expected %q, found %q", tc.scraper, tc.dateStr, res.DateStr)
		}
	}

}
