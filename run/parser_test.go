package run

import (
	"os"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func newDocumentFromFile(path string) (*goquery.Document, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return goquery.NewDocumentFromReader(f)
}

func TestFinanzaRepubblicaIt(t *testing.T) {
	path := "/home/mau/Code/go/src/github.com/mmbros/getstocks/data-crypt/doc/finanza.repubblica.it.html"
	doc, err := newDocumentFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	res, err := parseFinanzaRepubblicaIt(doc)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)

	expectedPriceStr := "90,680"
	expectedDateStr := "22/12/2016"
	if res.PriceStr != expectedPriceStr {
		t.Errorf("PriceStr: expected %q, found %q", expectedPriceStr, res.PriceStr)
	}
	if res.DateStr != expectedDateStr {
		t.Errorf("DateStr: expected %q, found %q", expectedDateStr, res.DateStr)
	}

}
