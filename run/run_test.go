package run

import (
	"bytes"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func parseDocA(doc *goquery.Document) (*ParseResult, error) {
	res := &ParseResult{
		DateStr: "OK",
	}
	return res, nil
}

func checkErr(t *testing.T, err error, prefix string) {
	if err == nil {
		t.Errorf("Expecting error with prefix %q, found no error.", prefix)
		return
	}
	if !strings.HasPrefix(err.Error(), prefix) {
		t.Errorf("Expecting error with prefix %q, found %q.", prefix, err.Error())
	}
}

func TestArgs(t *testing.T) {
	// nil scrapers -> error
	err := checkArgs(nil, nil)
	checkErr(t, err, "Scrapers must not be nil.")

	// scraper workers < 1 -> error
	scrapers := Scrapers{
		"SCRAPER_1": &Scraper{0, nil},
		"SCRAPER_2": &Scraper{2, parseDocA},
	}
	err = checkArgs(scrapers, nil)
	checkErr(t, err, "Invalid scraper: Workers must be > 0")

	// scraper ParseDoc == nil -> error
	scrapers["SCRAPER_1"].Workers = 1
	err = checkArgs(scrapers, nil)
	checkErr(t, err, "Invalid scraper: ParseDoc cannot be nil")

	// scraper ok
	scrapers["SCRAPER_1"].ParseDoc = parseDocA
	err = checkArgs(scrapers, nil)
	if err != nil {
		t.Errorf(err.Error())
	}

	// job with undefined scraper -> error
	jobs := Jobs{
		"JOB_1": []*JobReplica{&JobReplica{"SCRAPER_NOT_EXISTS", ""}},
	}
	err = checkArgs(scrapers, jobs)
	checkErr(t, err, "Invalid job: scraper key not found in scrapers")
}

func TestDispatcher(t *testing.T) {

	scrapers := Scrapers{
		"SITE_1": &Scraper{1, parseDocA},
		"SITE_2": &Scraper{2, parseDocA},
		"SITE_3": &Scraper{3, parseDocA},
	}
	jobs := Jobs{
		"JOB_1": []*JobReplica{
			&JobReplica{"SITE_1", ""},
			&JobReplica{"SITE_2", ""},
		},
		"JOB_2": []*JobReplica{
			&JobReplica{"SITE_2", ""},
			&JobReplica{"SITE_3", ""},
		},
		"JOB_3": []*JobReplica{
			&JobReplica{"SITE_1", ""},
			&JobReplica{"SITE_2", ""},
			&JobReplica{"SITE_3", ""},
		},
	}
	disp := newSimpleDispatcher(scrapers, jobs)

	buf := &bytes.Buffer{}
	disp.Debug(buf)
	t.Log(buf.String())
}
