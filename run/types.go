package run

import (
	"fmt"
	"io"
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

type dispatchItem struct {
	jobKey     JobKey
	jobReplica *JobReplica
}

type dispatchItems []*dispatchItem

type dispatcher map[*Scraper]dispatchItems

//-----------------------------------------------------------------------------

func (d dispatcher) Debug(w io.Writer) {
	fmt.Fprintln(w, "DISPATCHER")
	for _, items := range d {
		fmt.Fprintf(w, "Scraper %q\n", items[0].jobReplica.ScraperKey)
		for i, item := range items {
			fmt.Fprintf(w, "  [%d] %q\n", i, item.jobKey)
		}
	}
}

func newSimpleDispatcher(scrapers Scrapers, jobs Jobs) dispatcher {

	d := dispatcher(map[*Scraper]dispatchItems{})

	for key, replicas := range jobs {

		for _, replica := range replicas {
			item := &dispatchItem{
				jobKey:     key,
				jobReplica: replica,
			}
			scr := scrapers[replica.ScraperKey]
			items := d[scr]
			if items == nil {
				d[scr] = dispatchItems{item}
				continue
			}
			d[scr] = append(items, item)
		}

	}

	return d
}
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
