package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmbros/getstocks/cli"
	"github.com/mmbros/getstocks/run"
)

func main2() {
	os.Exit(cli.Run())
}

func parseDocA(*goquery.Document) (*run.ParseResult, error) {
	res := &run.ParseResult{
		PriceStr: "100.0",
		DateStr:  "oggi",
	}
	return res, nil
}

func main() {

	scrapers := run.Scrapers{
		"SITE_1": &run.Scraper{Workers: 1, ParseDoc: parseDocA},
		"SITE_2": &run.Scraper{Workers: 2, ParseDoc: parseDocA},
		"SITE_3": &run.Scraper{Workers: 3, ParseDoc: parseDocA},
	}
	jobs := run.Jobs{
		"JOB_1": []*run.JobReplica{
			&run.JobReplica{ScraperKey: "SITE_1", URL: ""},
			&run.JobReplica{ScraperKey: "SITE_2", URL: ""},
		},
		"JOB_2": []*run.JobReplica{
			&run.JobReplica{ScraperKey: "SITE_2", URL: ""},
			&run.JobReplica{ScraperKey: "SITE_3", URL: ""},
		},
		"JOB_3": []*run.JobReplica{
			&run.JobReplica{ScraperKey: "SITE_1", URL: ""},
			&run.JobReplica{ScraperKey: "SITE_2", URL: ""},
			&run.JobReplica{ScraperKey: "SITE_3", URL: ""},
		},
	}
	//disp := run.NewSimpleDispatcher(scrapers, jobs)

	//buf := &bytes.Buffer{}
	//disp.Debug(buf)
	//fmt.Println(buf.String())

	//disp = disp.Shuffle()
	//disp.Debug(buf)
	//fmt.Println(buf.String())

	ctx := context.Background()
	c, err := run.Execute(ctx, scrapers, jobs)
	if err != nil {
		fmt.Println(err)
		return
	}
	for item := range c {
		log.Printf("RESULT  %v\n", item)
	}

}
