package run

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func (di dispatchItems) scraperKey() ScraperKey {
	// NOTE: - at least one element mus be present
	//       - all the items must have the same ScraperKey
	return di[0].scraperKey
}

func (disp dispatcher) Debug(w io.Writer) {
	fmt.Fprintln(w, "DISPATCHER")
	for _, items := range disp {
		fmt.Fprintf(w, "Scraper %q\n", items.scraperKey())
		for i, item := range items {
			fmt.Fprintf(w, "  [%d] %q\n", i, item.jobKey)
		}
	}
}

func (disp dispatcher) Shuffle() dispatcher {
	shuffleItems := func(src dispatchItems) dispatchItems {
		dest := make(dispatchItems, len(src))
		perm := rand.Perm(len(src))
		log.Printf("%v\n", perm)
		for i, v := range perm {
			dest[v] = src[i]
		}
		return dest
	}

	d := dispatcher(map[*Scraper]dispatchItems{})

	for scraper, items := range disp {
		d[scraper] = shuffleItems(items)
	}

	return d

}
func NewSimpleDispatcher(scrapers Scrapers, jobs Jobs) dispatcher {

	d := dispatcher(map[*Scraper]dispatchItems{})

	for key, replicas := range jobs {

		for _, replica := range replicas {
			item := &dispatchItem{
				jobKey:     key,
				scraperKey: replica.ScraperKey,
				url:        replica.URL,
			}
			scr := scrapers[replica.ScraperKey]
			items := d[scr]
			if items == nil {
				d[scr] = dispatchItems{item}
				continue
			}
			d[scr] = append(items, item)
		}

	}

	return d
}

func checkArgs(scrapers Scrapers, jobs Jobs) error {
	if scrapers == nil {
		return fmt.Errorf("Scrapers must not be nil.")
	}

	// check jobs
	for jk, replicas := range jobs {
		if len(replicas) < 1 {
			return fmt.Errorf("Invalid job: no replica found (job %q).", jk)
		}
		for ri, rv := range replicas {
			// check replica != nil
			if rv == nil {
				return fmt.Errorf("Invalid job: replica cannot be nil (job %q, replica #%d).", jk, ri)
			}
			// check scraperKey exists in scrapers
			if _, ok := scrapers[rv.ScraperKey]; !ok {
				return fmt.Errorf("Invalid job: scraper key not found in scrapers (job %q, replica #%d, scraper %q).", jk, ri, rv.ScraperKey)
			}
		}
	}

	// check scrapers
	for sk, sv := range scrapers {
		if sv.Workers <= 0 {
			return fmt.Errorf("Invalid scraper: Workers must be > 0 (scraper %q).", sk)
		}
		if sv.ParseDoc == nil {
			return fmt.Errorf("Invalid scraper: ParseDoc cannot be nil (scraper %q).", sk)
		}
	}

	return nil
}

func genWorkRequestChan(jobContexts map[JobKey]*jobContext, items dispatchItems) chan *workRequest {
	out := make(chan *workRequest)
	go func() {
		for _, item := range items {
			jc := jobContexts[item.jobKey]
			req := &workRequest{
				ctx:     jc.ctx,
				resChan: jc.resChan,
				item:    item,
			}
			out <- req
		}
		close(out)
	}()
	return out
}

// Execute starts the execution of the jobs dispatching each job replica
// to the corrisponding scraper.
func Execute(ctx context.Context, scrapers Scrapers, jobs Jobs) (chan *WorkResult, error) {
	if jobs == nil || len(jobs) == 0 {
		// nothing to do!
		return nil, nil
	}
	// check args
	if err := checkArgs(scrapers, jobs); err != nil {
		return nil, err
	}

	// build dispatcher
	disp := NewSimpleDispatcher(scrapers, jobs).Shuffle()

	// create a context with cancel and a result chan for each enabled stock
	jobContexts := map[JobKey]*jobContext{}
	for jobKey, jobReplicas := range jobs {
		ctx0, cancel0 := context.WithCancel(ctx)
		jobContexts[jobKey] = &jobContext{
			ctx:     ctx0,
			cancel:  cancel0,
			resChan: make(chan *WorkResult, len(jobReplicas)),
		}

	}

	// create a request chan for each scraper and enqueues the jobs
	reqChan := map[*Scraper]chan *workRequest{}
	for scraper, items := range disp {
		reqChan[scraper] = genWorkRequestChan(jobContexts, items)
	}

	out := make(chan *WorkResult)

	// raccoglie le risposte per ogni job
	var wg sync.WaitGroup
	wg.Add(len(jobs))
	for jobKey, replicas := range jobs {

		go func(jobctx *jobContext, count int) {
			todo := true

			for ; count > 0; count-- {

				select {
				case res := <-jobctx.resChan:

					// if not already done, send the result if ok,
					// or if it is the last result.
					if todo && (res.Err == nil || count == 1) {
						todo = false
						jobctx.cancel()
						out <- res
					}
				case <-jobctx.ctx.Done():
					jobctx.cancel()
				}

			}
			wg.Done()
		}(jobContexts[jobKey], len(replicas))
	}
	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		log.Println("CLOSING OUT")
		close(out)
	}()

	// crea le istanze dei workers che lavorano i jobs
	for scraper, items := range disp {
		for j := 0; j < scraper.Workers; j++ {
			worker := newScraperWorker(items.scraperKey(), scraper, j+1)

			go func(w *scraperWorker, input <-chan *workRequest) {
				// per ogni work request ottenuto dal chan
				for req := range input {
					req.resChan <- w.doJob(req.ctx, req.item)
				}

			}(worker, reqChan[scraper])
		}
	}
	return out, nil
}

func newScraperWorker(key ScraperKey, scraper *Scraper, index int) *scraperWorker {
	return &scraperWorker{
		key:     key,
		scraper: scraper,
		index:   index,
	}
}

func (sw *scraperWorker) String() string {
	if sw == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s-%d", sw.key, sw.index)
}

func (w *scraperWorker) doJob(ctx context.Context, job *dispatchItem) (result *WorkResult) {
	log.Printf("JOB IN  %s - %s", w, job)
	// init the result
	result = &WorkResult{
		ScraperKey: job.scraperKey,
		JobKey:     job.jobKey,
		URL:        job.url,
		TimeStart:  time.Now(),
	}
	// use defer to set timeEnd
	defer func() {
		result.TimeEnd = time.Now()

		log.Printf("JOB OUT %s - %s, err:%v ", w, job, result.Err)
	}()

	// get the response
	resp, err := GetUrl(ctx, job.url)
	if err != nil {
		result.Err = err
		return
	}
	// check response status
	if resp.StatusCode != http.StatusOK {
		result.Err = fmt.Errorf(resp.Status)
		return
	}
	// create goquery document
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		result.Err = err
		return
	}
	// parse and return the response
	result.Res, result.Err = w.scraper.ParseDoc(doc)

	return
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
