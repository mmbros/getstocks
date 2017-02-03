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

type parseDocFunc func(doc *goquery.Document) (*parseResult, error)

func getParseDocFunc(scraperName string) parseDocFunc {
	var m = map[string]parseDocFunc{
		"finanza.repubblica.it": parseFinanzaRepubblicaIt,
		"www.borse.it":          parseWwwBorseIt,
		"www.eurotlx.com":       parseWwwEurotlxCom,
		"www.milanofinanza.it":  parseWwwMilanofinanzaIt,
		"www.morningstar.it":    parseWwwMorningstarIt,
		"www.teleborsa.it":      parseWwwTeleborsaIt,
	}
	return m[scraperName]
}

// ============================================================================

func parseDate(layout, str string) (time.Time, error) {
	var t time.Time
	if str == "" {
		return t, errors.New("Date not found")
	}
	loc, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		return t, err
	}
	return time.ParseInLocation(layout, str, loc)
}
func parsePrice(str string) (float32, error) {
	if str == "" {
		return 0.0, errors.New("Price not found")
	}
	price, err := strconv.ParseFloat(strings.Replace(str, ",", ".", 1), 32)
	return float32(price), err
}

func (pr *parseResult) setPriceAndDate(layout string) error {
	var err, err2 error
	pr.Price, err = parsePrice(pr.PriceStr)
	pr.Date, err2 = parseDate(layout, pr.DateStr)

	if err == nil {
		err = err2
	}
	return err
}

// ============================================================================

func parseFinanzaRepubblicaIt(doc *goquery.Document) (*parseResult, error) {
	res := &parseResult{}

	doc.Find("div.TLB-scheda-body-container > ul > li:first-child > b ").EachWithBreak(func(i int, s *goquery.Selection) bool {
		switch i {
		case 2:
			res.PriceStr = s.Text()
		case 3:
			res.DateStr = s.Text()
			return false
		}
		return true
	})

	if err := res.setPriceAndDate("02/01/2006"); err != nil {
		return nil, err
	}
	return res, nil
}

/*
  <table>
    <tr>
      <th colspan="2">Dati giornalieri</th>
    </tr>
    <tr>
      <td class="table_label">Prezzo di chiusura</td>
      <td>90,68</td>
    </tr>
    <tr>
      <td class="table_label">Data</td>
      <td>30-01-2017</td>
    </tr>
*/
func parseWwwEurotlxCom(doc *goquery.Document) (*parseResult, error) {
	res := &parseResult{}

	doc.Find("td.table_label").EachWithBreak(func(i int, s *goquery.Selection) bool {

		if s.Text() == "Prezzo di chiusura" {
			sel := s.Next()
			res.PriceStr = sel.Text()
			sel = s.Parent().Next().Children().First().Next()
			res.DateStr = sel.Text()
			return false
		}
		return true
	})

	if err := res.setPriceAndDate("02-01-2006"); err != nil {
		return nil, err
	}
	return res, nil
}

//<div class="fleft mright10"><span><span class="info font22 ">113,19</span>
//<div class="fleft w65 taright bold"><span class="cred font12 taright">-0,0618</span>
//<div class="mtop10 bgees mbottom5"><span class="cred"> 03/02/17 18.02.03 </span>

func parseWwwMilanofinanzaIt(doc *goquery.Document) (*parseResult, error) {
	res := &parseResult{}

	res.PriceStr = doc.Find(".font22").Text()
	res.DateStr = strings.TrimSpace(doc.Find("div.mbottom5 span.cred").Text())

	if err := res.setPriceAndDate("02/01/06 15.04.05"); err != nil {
		return nil, err
	}
	return res, nil
}

//<div id="quotazioni" style="float: left"></div>
//<div class="schede">
//<ul>
//<li class="titolo">Chiusura</li>
//<li class="descr przAcq">5,3600</li>
//<li class="titolo">Var. %</li>
//<li class="descr przAcq">-0,060%</li>
//<li class="titolo">Società di gestione</li>
//<li class="descr przAcq">Anima Sgr Spa<li>
//<li class="titolo">Isin</li>
//<li class="descr">&nbsp;IT0004930167</li>
//</ul>
//<ul>
//<li class="titolo">Data</li>
//<li class="descr Data">20/01/2017</li>
//<li class="titolo">Valuta</li>
//<li class="descr przAcq">EUR</li>
//<li class="titolo">Tipologia</li>
//<li class="descr przAcq">F. Comuni<li>
//</ul>
func parseWwwBorseIt(doc *goquery.Document) (*parseResult, error) {
	res := &parseResult{}

	doc.Find("div.schede > ul > li.descr").EachWithBreak(func(i int, s *goquery.Selection) bool {
		switch i {
		case 0:
			res.PriceStr = s.Text()
		case 4:
			res.DateStr = s.Text()
			return false
		}
		return true
	})

	if err := res.setPriceAndDate("02/01/2006"); err != nil {
		return nil, err
	}
	return res, nil
}

//<div id="ctl00_phContents_ctlHeader_pnlHeaderContainer" class="panel-header-scheda">
//
//<div id="ctl00_phContents_ctlHeader_pnlHeaderMarketInfo" class="header-market-info fc2">
//
//ISIN: IT0004930167 - Mercato: Fondi e SICAV
//
//</div>
//<div id="ctl00_phContents_ctlHeader_pnlHeaderTop" class="header-top title fc1">
//
//<span id="ctl00_phContents_ctlHeader_lblPercentChange" class="h-trend aumento arrow-aumento">+0,15%</span>
//<span id="ctl00_phContents_ctlHeader_lblPrice" class="h-price fc0">5,368</span>
//<h1>Anima Traguardo 2019 Plus II</h1>
//
//</div>
//
//<div id="ctl00_phContents_ctlHeader_pnlHeaderBottom" class="header-bottom fc3">
//
//
//Ultimo aggiornamento: <strong>27/01/2017</strong>
//
//</div>
//
//</div>
func parseWwwTeleborsaIt(doc *goquery.Document) (*parseResult, error) {
	res := &parseResult{}

	res.PriceStr = doc.Find("#ctl00_phContents_ctlHeader_lblPrice").Text()
	res.DateStr = doc.Find("#ctl00_phContents_ctlHeader_pnlHeaderBottom strong").Text()

	if err := res.setPriceAndDate("02/01/2006"); err != nil {
		return nil, err
	}
	return res, nil
}

// <table class="snapshotTextColor snapshotTextFontStyle snapshotTable overviewKeyStatsTable" border="0">
//   <tr><td class="titleBarHeading" colspan="3">Sintesi</td></tr>
//   <tr><td class="line heading">NAV<span class="heading"><br />27/01/2017</span></td>
//       <td class="line"> </td>
//       <td class="line text">EUR 5,158</td></tr>
//   <tr><td class="line heading">Var.Ultima Quotazione</td><td class="line"> </td><td class="line text">-0,04%
func parseWwwMorningstarIt(doc *goquery.Document) (*parseResult, error) {
	res := &parseResult{}

	doc.Find("table.overviewKeyStatsTable td").EachWithBreak(func(i int, s *goquery.Selection) bool {
		switch i {
		case 1:
			res.DateStr = s.Find("span").Text()
		case 3:
			res.PriceStr = s.Text()[5:] // drop start "EUR "
			return false
		}
		return true
	})

	if err := res.setPriceAndDate("02/01/2006"); err != nil {
		return nil, err
	}
	return res, nil
}
