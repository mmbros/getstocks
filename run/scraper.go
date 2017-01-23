package run

import (
	"errors"
	"fmt"
)

// must be >= 1
const maxScraperWorkers = 10

type Scraper struct {
	scrapertype ScraperType
	workers     int
}

type Scrapers struct {
	m map[ScraperType]int
}

func NewScrapers() *Scrapers {
	return &Scrapers{
		m: map[ScraperType]int{},
	}
}
func (scrs *Scrapers) Add(t ScraperType, workers int) error {
	if scrs == nil {
		return errors.New("Scrapers in nil")
	}
	if fn := t.getParseDocFunc(); fn == nil {
		return fmt.Errorf("Invalid scraper: no parse doc function: %q", t)
	}
	if _, ok := scrs.m[t]; ok {
		return fmt.Errorf("Scraper already exists: %q", t)
	}
	if workers < 1 {
		workers = 1
	}
	if workers > maxScraperWorkers {
		workers = maxScraperWorkers
	}

	scrs.m[t] = workers
	return nil

}

func NewScraper(scrapertype ScraperType, workers int) (*Scraper, error) {

	if fn := scrapertype.getParseDocFunc(); fn == nil {
		return nil, fmt.Errorf("Invalid scraper: no parse doc function: %q", scrapertype)
	}
	if workers < 1 {
		workers = 1
	}
	if workers > maxScraperWorkers {
		workers = maxScraperWorkers
	}

	return &Scraper{
		scrapertype: scrapertype,
		workers:     workers,
	}, nil
}

func (s Scraper) ScraperType() ScraperType { return s.scrapertype }
func (s Scraper) Workers() int             { return s.workers }
