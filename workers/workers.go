package workers

import (
	"context"
	"fmt"
	"sync"
)

// Max number of instances for each worker
const maxInstances = 100

//-----------------------------------------------------------------------------
// Types to be customized if needed
//-----------------------------------------------------------------------------

// WorkerKey type definition.
type WorkerKey string

// JobKey type definition.
type JobKey string

//-----------------------------------------------------------------------------

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

//-----------------------------------------------------------------------------

type jobContextItem struct {
	workers int
	ctx     context.Context
	cancel  context.CancelFunc
	resChan chan Response
}

type dispatcher struct {
	workerMap      map[WorkerKey]*Worker
	workerRequests mapWorkerRequests
	jobContext     map[JobKey]*jobContextItem
}

type task struct {
	ctx     context.Context
	req     Request
	resChan chan Response
}

//func (d *dispatcher) jobs() int {
//return len(d.jobContext)
//}

//func (d *dispatcher) workers() int {
//return len(d.workerQueue)
//}

func newBasicDispatcher(ctx context.Context, workers []*Worker, requests []Request) (*dispatcher, error) {

	// check workers and map form workerid to Worker
	wm := map[WorkerKey]*Worker{}
	for _, w := range workers {
		if _, ok := wm[w.WorkerID]; ok {
			return nil, fmt.Errorf("Duplicate worker: %q", w.WorkerID)
		}
		if w.Instances <= 0 || w.Instances > maxInstances {
			return nil, fmt.Errorf("Instances must be in 1..%d range: worker=%q", maxInstances, w.WorkerID)
		}
		if w.Work == nil {
			return nil, fmt.Errorf("Work function cannot be nil: worker=%q", w.WorkerID)
		}
		wm[w.WorkerID] = w
	}

	d := &dispatcher{
		workerMap:      wm,
		workerRequests: mapWorkerRequests{},
		jobContext:     map[JobKey]*jobContextItem{},
	}

	// check requests and init dispatcher
	for _, r := range requests {

		// updates worker queue
		wid := r.WorkerID()
		// get the worker queue, if exists
		if wr, ok := d.workerRequests[wid]; ok {
			// append the new request to the queue
			d.workerRequests[wid] = append(wr, r)
		} else {
			// check if the worker exists
			if _, ok := wm[wid]; !ok {
				return nil, fmt.Errorf("Worker not found: worker=%q, job=%q", wid, r.JobID())
			}
			// create the worker queue
			d.workerRequests[wid] = []Request{r}
		}

		// updates job contexts
		jid := r.JobID()
		jc := d.jobContext[jid]
		if jc == nil {
			jc = &jobContextItem{}
			d.jobContext[jid] = jc
			// create the job context and cancel function
			jc.ctx, jc.cancel = context.WithCancel(ctx)
			// in case of buffered chan, can't create here the resChan
			//jc.resChan = make(chan Response) // see below
		}
		// NOTE: doesn't check if the worker has already been used for the same job
		jc.workers = jc.workers + 1

	}

	// create the resChan buffered channel for each job
	for _, jc := range d.jobContext {
		jc.resChan = make(chan Response, jc.workers)
	}

	return d, nil
}

// genTaskChan returns a chan where are enqueued the task for the worker
func (d *dispatcher) genTaskChan(wid WorkerKey) chan *task {
	out := make(chan *task)
	go func() {
		for _, item := range d.workerRequests[wid] {
			jc := d.jobContext[item.JobID()]
			wreq := &task{
				ctx:     jc.ctx,
				resChan: jc.resChan,
				req:     item,
			}
			out <- wreq
		}
		close(out)
	}()
	return out
}

func (jc *jobContextItem) getJobResponse(out chan Response) {
	todo := true
	count := jc.workers

	for ; count > 0; count-- {

		select {
		case res := <-jc.resChan:
			// if not already done,
			// send the result if Success,
			// or if it is the last result.
			if todo && (res.Success() || count == 1) {
				todo = false
				jc.cancel()
				out <- res
				// XXX: inserted return statement: check it!!!
				//return
			}
		case <-jc.ctx.Done():
			jc.cancel()
		}
	}
}

// Execute function is ...
func Execute(ctx context.Context, workers []*Worker, requests []Request) (chan Response, error) {

	// Create the dispatcher
	d, err := newBasicDispatcher(ctx, workers, requests)
	if err != nil {
		return nil, err
	}
	//d.workerRequests.shuffle()
	d.workerRequests.distribute()

	// Generate a task chan for each worker
	reqChan := map[WorkerKey]chan *task{}
	for wid := range d.workerRequests {
		reqChan[wid] = d.genTaskChan(wid)
	}

	// Creates the output channel
	out := make(chan Response)

	// Starts a goroutine for each job to wait for the job response
	var wg sync.WaitGroup
	wg.Add(len(d.jobContext))
	for _, jc := range d.jobContext {
		go func(jc *jobContextItem) {
			jc.getJobResponse(out)
			wg.Done()
		}(jc)
	}
	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		//log.Println("CLOSING OUT")
		close(out)
	}()

	// Starts the goroutines that executes the real work.
	// For each worker it starts N goroutines, with N = Instances.
	// Each goroutine get the input request from the worker request channel,
	// and put the output response to the job response channel.
	for wid, worker := range d.workerMap {
		// get the request channel of the worker
		reqc := reqChan[wid]
		// for each worker instances
		for i := 0; i < worker.Instances; i++ {

			go func(w *Worker, input <-chan *task) {
				for wreq := range input {
					wreq.resChan <- w.Work(wreq.ctx, wreq.req)
				}
			}(worker, reqc)

		}
	}

	return out, nil
}
