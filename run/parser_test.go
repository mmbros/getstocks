package run

import (
	"os"
	"path/filepath"
	"strings"
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
		filename string
		priceStr string
		dateStr  string
	}{
		{"finanza.repubblica.it", "finanza.repubblica.it.html", "90,680", "22/12/2016"},
		{"www.borse.it", "www.borse.it.html", "5,3600", "20/01/2017"},
		{"www.eurotlx.com", "www.eurotlx.com.html", "90,68", "30-01-2017"},
		{"www.milanofinanza.it", "www.milanofinanza.it.IT0004009673.html", "113,19", "03/02/17 18.02.03"},
		{"www.milanofinanza.it", "www.milanofinanza.it.IT0004977085.html", "5,052", "27/01/17 1.00.00"},
		{"www.milanofinanza.it", "www.milanofinanza.it.html", "5,048", "20/01/17 1.00.00"},
		{"www.morningstar.it", "www.morningstar.it.html", "5,158", "27/01/2017"},
		{"www.teleborsa.it", "www.teleborsa.it.html", "5,368", "27/01/2017"},
		{"www.mpscapitalservices.it", "www.mpscapitalservices.it.html", "103.58", "30/01/2017"},
	}
	for _, tc := range testCases {

		path := getpath(tc.filename)
		doc, err := newDocumentFromFile(path)
		if err != nil {
			t.Error(tc.filename, err)
			continue
		}
		parseFunc := getParseDocFunc(tc.scraper)
		res, err := parseFunc(doc)
		if err != nil {
			t.Error(tc.filename, err)
			continue
		}
		t.Log(tc.filename, "->", res)

		if res.PriceStr != tc.priceStr {
			t.Errorf("[%s] PriceStr: expected %q, found %q", tc.filename, tc.priceStr, res.PriceStr)
		}
		if res.DateStr != tc.dateStr {
			t.Errorf("[%s] DateStr: expected %q, found %q", tc.filename, tc.dateStr, res.DateStr)
		}
	}

}
func TestParserRule(t *testing.T) {
	var filename = "www.mpscapitalservices.it.html"
	var s string

	path := getpath(filename)
	doc, err := newDocumentFromFile(path)
	if err != nil {
		t.Fatal(filename, err)
	}

	s = doc.Find("#Official td.top").Text()
	t.Logf("PriceStr: %s", s)

	s = doc.Find("#lbMercato").Parent().Text()
	s = strings.TrimSpace(s)
	s = s[len(s)-10:]

	t.Logf("DateStr: %s", s)
}
