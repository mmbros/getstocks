package run

import (
	"context"
	"time"

	"github.com/PuerkitoBio/goquery"
)

//-----------------------------------------------------------------------------
// Types to be customized if needed
//-----------------------------------------------------------------------------

// ScraperKey type definition.
type ScraperKey string

// JobKey type definition.
type JobKey string

// ParseResult represents the type of the value returned by the ParseDoc function.
type ParseResult struct {
	Price    float32
	Date     time.Time
	PriceStr string
	DateStr  string
}

func (pr *ParseResult) String() string {
	if pr == nil {
		return "<nil>"
	}
	return pr.PriceStr
}

//-----------------------------------------------------------------------------
// Static types
//-----------------------------------------------------------------------------

// ParseDocFunc ...
type ParseDocFunc func(*goquery.Document) (*ParseResult, error)

// Scraper ...
type Scraper struct {
	Workers  int
	ParseDoc ParseDocFunc
}

// Scrapers ...
type Scrapers map[ScraperKey]*Scraper

// JobReplica ...
type JobReplica struct {
	ScraperKey ScraperKey
	URL        string
}

// Jobs ...
type Jobs map[JobKey][]*JobReplica

//-----------------------------------------------------------------------------
// private types
//-----------------------------------------------------------------------------
type dispatchItem struct {
	scraperKey ScraperKey
	jobKey     JobKey
	url        string
}

type dispatchItems []*dispatchItem

type dispatcher map[*Scraper]dispatchItems

type workRequest struct {
	ctx     context.Context
	resChan chan *WorkResult
	item    *dispatchItem
}

type scraperWorker struct {
	key     ScraperKey
	scraper *Scraper
	index   int
}

type WorkResult struct {
	ScraperKey ScraperKey
	JobKey     JobKey
	URL        string
	TimeStart  time.Time
	TimeEnd    time.Time
	Res        *ParseResult
	Err        error
}

func (wr *WorkResult) Elapsed() time.Duration {
	return wr.TimeEnd.Sub(wr.TimeStart)
}

type jobContext struct {
	ctx     context.Context
	cancel  context.CancelFunc
	resChan chan *WorkResult
}
