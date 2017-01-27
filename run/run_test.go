package run

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

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
		"SCRAPER_2": &Scraper{2, parseDocTest},
	}
	err = checkArgs(scrapers, nil)
	checkErr(t, err, "Invalid scraper: Workers must be > 0")

	// scraper ParseDoc == nil -> error
	scrapers["SCRAPER_1"].Workers = 1
	err = checkArgs(scrapers, nil)
	checkErr(t, err, "Invalid scraper: ParseDoc cannot be nil")

	// scraper ok
	scrapers["SCRAPER_1"].ParseDoc = parseDocTest
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

		"SITE_1": &Scraper{1, parseDocTest},
		"SITE_2": &Scraper{2, parseDocTest},
		"SITE_3": &Scraper{3, parseDocTest},
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

	disp = disp.shuffle()
	t.Log(buf.String())
}

func TestExecute(t *testing.T) {

	server := initTestServer()

	jobrep := func(site string, msecMin, msecMax, err int) *JobReplica {
		u, _ := url.Parse(server.URL)
		q := u.Query()
		q.Set("msec1", strconv.Itoa(msecMin))
		q.Set("msec2", strconv.Itoa(msecMax))
		q.Set("err", strconv.Itoa(err))
		u.RawQuery = q.Encode()
		u.Path = site

		return &JobReplica{
			ScraperKey: ScraperKey(site),
			URL:        u.String(),
		}
	}
	jobrep0 := func(site string) *JobReplica {
		return jobrep(site, 100, 100, 50)
	}

	scrapers := Scrapers{
		"SITE_1": &Scraper{1, parseDocTest},
		"SITE_2": &Scraper{2, parseDocTest},
		"SITE_3": &Scraper{3, parseDocTest},
	}

	jobs := Jobs{
		"JOB_1": []*JobReplica{
			jobrep0("SITE_1"),
			jobrep0("SITE_2"),
		},
		"JOB_2": []*JobReplica{
			jobrep0("SITE_2"),
			jobrep0("SITE_3"),
		},
		"JOB_3": []*JobReplica{
			jobrep0("SITE_1"),
			jobrep0("SITE_2"),
			jobrep0("SITE_3"),
		},
		"JOB_4": []*JobReplica{
			jobrep0("SITE_1"),
			jobrep0("SITE_3"),
		},
		"JOB_5": []*JobReplica{
			jobrep0("SITE_1"),
			jobrep0("SITE_2"),
		},
		"JOB_6": []*JobReplica{
			jobrep0("SITE_2"),
		},
		"JOB_7": []*JobReplica{
			jobrep0("SITE_1"),
			jobrep0("SITE_2"),
		},
		"JOB_8": []*JobReplica{
			jobrep0("SITE_3"),
		},
	}

	t1 := time.Now()
	ctx := context.Background()
	c, err := Execute(ctx, scrapers, jobs)
	if err != nil {
		t.Error(err.Error())
		return
	}
	for item := range c {
		t.Logf("RESULT  %d %s  %s  %v  %v\n", item.Elapsed().Nanoseconds()/1000000, item.JobKey, item.ScraperKey, item.Res, item.Err)
	}
	t2 := time.Now()
	fmt.Printf("*** ELAPSED %s\n", t2.Sub(t1))
}
