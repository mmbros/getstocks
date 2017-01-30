package run

import (
	"context"
	"net/http"
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

func GetScraperFromUrl(url string) (string, error) {
	return "", nil
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
