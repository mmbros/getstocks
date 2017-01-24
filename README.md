# getstocks
retrieve stocks quotes from web sites


## getstocks/run

### type definitions

Tipi da modificare in funzione degli scopi
- ScraperKey: string
- JobKey: string
- Result: {price:float32, date:time.Time}

Tipi indipendenti dallo scopo particolare
- ParseDocFunc: func(doc *goquery.Document) (*Result, error)
- Scrapers: map[key: ScraperKey] struct { workers: int, fn ParseDocFunc}
- Jobs: map[key: JobId] []*JobReplica
- JobReplica: struct{scraper: ScraperId, url: string}

### func Execute

checks args:

- Jobs:
  - *JobReplica != nil
  - scraper deve corrispondere ad uno degli scrapers
  - len(JobReplica) >= 1

- Scrapers:
  - workers > 0
  - fn not nil


