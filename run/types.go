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
	dispatchItem
	Res       *ParseResult
	TimeStart time.Time
	TimeEnd   time.Time
	Err       error
}

type jobContext struct {
	ctx     context.Context
	cancel  context.CancelFunc
	resChan chan *WorkResult
}
