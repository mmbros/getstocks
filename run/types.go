package run

import (
	"fmt"
	"time"

	"github.com/PuerkitoBio/goquery"
)

//-----------------------------------------------------------------------------
// generics types
//-----------------------------------------------------------------------------
type ScraperKey string

type JobKey string

type ParseResult struct {
	Price    float32
	Date     time.Time
	PriceStr string
	DateStr  string
}

//-----------------------------------------------------------------------------
// static types
//-----------------------------------------------------------------------------
type ParseDocFunc func(*goquery.Document) (*ParseResult, error)

type Scraper struct {
	Workers  int
	ParseDoc ParseDocFunc
}

type Scrapers map[ScraperKey]*Scraper

type JobReplica struct {
	ScraperKey ScraperKey
	URL        string
}

type Jobs map[JobKey][]*JobReplica

//-----------------------------------------------------------------------------

func checkArgs(scrapers Scrapers, jobs Jobs) error {
	if scrapers == nil {
		return fmt.Errorf("Scrapers must not be nil.")
	}

	// check jobs
	for jk, replicas := range jobs {
		if len(replicas) < 1 {
			return fmt.Errorf("Invalid job: no replica found (job %q).", jk)
		}
		for ri, rv := range replicas {
			// check replica != nil
			if rv == nil {
				return fmt.Errorf("Invalid job: replica cannot be nil (job %q, replica #%d).", jk, ri)
			}
			// check scraperKey exists in scrapers
			if _, ok := scrapers[rv.ScraperKey]; !ok {
				return fmt.Errorf("Invalid job: scraper key not found in scrapers (job %q, replica #%d, scraper %q).", jk, ri, rv.ScraperKey)
			}
		}
	}

	// check scrapers
	for sk, sv := range scrapers {
		if sv.Workers <= 0 {
			return fmt.Errorf("Invalid scraper: Workers must be > 0 (scraper %q).", sk)
		}
		if sv.ParseDoc == nil {
			return fmt.Errorf("Invalid scraper: ParseDoc cannot be nil (scraper %q).", sk)
		}
	}

	return nil
}

func Execute(scrapers Scrapers, jobs Jobs) error {
	if jobs == nil || len(jobs) == 0 {
		// nothing to do!
		return nil
	}
	// check args
	if err := checkArgs(scrapers, jobs); err != nil {
		return err
	}
	return nil
}
