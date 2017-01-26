package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/mmbros/getstocks/cli"
	"github.com/mmbros/getstocks/run"
)

func main2() {
	os.Exit(cli.Run())
}

func main() {
	parseDocA := run.ParseDocFunc(nil)

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
	disp := run.NewSimpleDispatcher(scrapers, jobs)

	buf := &bytes.Buffer{}
	disp.Debug(buf)
	fmt.Println(buf.String())

	disp = disp.Shuffle()
	disp.Debug(buf)
	fmt.Println(buf.String())

}
