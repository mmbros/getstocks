package run

import (
	"context"
	"fmt"
	"net/http"
	neturl "net/url"
	"time"
)

type Scraper struct {
	Name    string
	Workers int
}

type Stock struct {
	Name        string
	Isin        string
	Description string
	Sources     []*StockSource
}

type StockSource struct {
	Scraper string
	URL     string
}

// ----------------------------------------------------------------------------
/*
// Request inteface.
type Request interface {
	JobID() JobKey
	WorkerID() WorkerKey
}

// Response is the interface that must be matched by the results of the Work function.
type Response interface {
	// Success return true in case of a success response.
	// In this case no other Request will be worked for the same Job.
	Success() bool
}

// WorkFunc is the worker function.
type WorkFunc func(context.Context, Request) Response

// Worker is ...
type Worker struct {
	WorkerID  WorkerKey
	Instances int
	Work      WorkFunc
}
*/
type request struct {
	scraperName string
	stockName   string
	URL         string
}

func (req *request) WorkerID() string { return req.scraperName }
func (req *request) JobID() string    { return req.stockName }

type Response struct {
	ScraperName string
	StockName   string
	URL         string
	PriceStr    string
	DateStr     string
	Price       float32
	Date        time.Time
	Err         error
}

func (res *Response) Success() bool { return res.Err == nil }

// ----------------------------------------------------------------------------

func GetScraperFromUrl(url string) (string, error) {
	// get the host form url
	u, err := neturl.Parse(url)
	if err != nil {
		return "", err
	}
	name := u.Host
	// check if to the name corresponds a ParseDocFunc
	if getParseDocFunc(name) == nil {
		return "", fmt.Errorf("No scraper found for url %q", url)
	}
	return name, nil
}

func GetUrl(ctx context.Context, url string) (*http.Response, error) {

	type result struct {
		resp *http.Response
		err  error
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// make the request
		tr := &http.Transport{}
		client := &http.Client{Transport: tr}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		c := make(chan result, 1)

		go func() {
			resp, err := client.Do(req)
			c <- result{resp: resp, err: err}
		}()

		select {
		case <-ctx.Done():
			tr.CancelRequest(req)
			<-c // Wait for client.Do
			return nil, ctx.Err()
		case r := <-c:
			return r.resp, r.err
		}
	}
}
