package run

import (
	"context"
	"fmt"
	"net/http"
	neturl "net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmbros/getstocks/workers"
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

func (req *request) WorkerID() workers.WorkerKey { return workers.WorkerKey(req.scraperName) }
func (req *request) JobID() workers.JobKey       { return workers.JobKey(req.stockName) }

type Response struct {
	ScraperName string
	StockName   string
	URL         string
	Result      *parseResult
	TimeStart   time.Time
	TimeEnd     time.Time
	Err         error
}

func (res *Response) Success() bool { return res.Err == nil }

// ----------------------------------------------------------------------------

func GetScraperFromUrl(url string) (string, error) {
	// Get the host form url
	u, err := neturl.Parse(url)
	if err != nil {
		return "", err
	}
	name := u.Host
	// Checks if to the name corresponds a ParseDocFunc.
	// It returns anyway the supposed name of the scraper.
	if getParseDocFunc(name) == nil {
		return name, fmt.Errorf("No scraper found for url %q", url)
	}
	return name, nil
}

func getUrl(ctx context.Context, url string) (*http.Response, error) {

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

//type WorkFunc func(context.Context, Request) Response
func scraperWorkFunc(ctx context.Context, wreq workers.Request) workers.Response {

	//workers.Request -> *request
	//workers.Response -> Response
	req := wreq.(*request)

	// init the result
	response := &Response{
		ScraperName: req.scraperName,
		StockName:   req.stockName,
		URL:         req.URL,
		TimeStart:   time.Now(),
	}
	// use defer to set timeEnd
	defer func() {
		response.TimeEnd = time.Now()
	}()

	// get the http response
	resp, err := getUrl(ctx, req.URL)
	if err != nil {
		response.Err = err
		return response
	}
	if resp.StatusCode != http.StatusOK {
		response.Err = fmt.Errorf(resp.Status)
		return response
	}

	// create goquery document
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		response.Err = err
		return response
	}
	// parse the response
	parseFunc := getParseDocFunc(req.scraperName)
	response.Result, err = parseFunc(doc)
	if err != nil {
		response.Err = err
		return response
	}

	return response

}

func Execute(ctx context.Context, scrapers []*Scraper, stocks []*Stock) (<-chan *Response, error) {

	// init list of workers.Worker
	wrks := make([]*workers.Worker, 0, len(scrapers))
	for _, scr := range scrapers {
		w := &workers.Worker{
			WorkerID:  workers.WorkerKey(scr.Name),
			Instances: scr.Workers,
			Work:      scraperWorkFunc,
		}
		wrks = append(wrks, w)
	}

	// Init list of workers.Request.
	// Assumes each stock has 3 sources.
	reqs := make([]workers.Request, 0, 3*len(stocks))
	for _, stock := range stocks {
		for _, src := range stock.Sources {
			r := &request{
				scraperName: src.Scraper,
				stockName:   stock.Name,
				URL:         src.URL,
			}
			reqs = append(reqs, r)
		}
	}

	// Call workers.Execute to do the job
	wout, err := workers.Execute(ctx, wrks, reqs)
	if err != nil {
		return nil, err
	}

	// Creates the output channel
	out := make(chan *Response)

	// Starts a goroutine that:
	// 1. get each workers.Response from the wout channel,
	// 2. traforms it to *run.Response type,
	// 3. sends it to the out channel.
	go func() {
		for wres := range wout {
			res := wres.(*Response)
			out <- res
		}
		close(out)
	}()

	return out, nil
}
