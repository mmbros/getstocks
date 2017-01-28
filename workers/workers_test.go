package workers

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

type testRequest struct {
	jobid    string
	workerid string
	minMsec  int
	maxMsec  int
	percErr  int
	t        *testing.T
}
type testResponse struct {
	jobid    string
	workerid string
	result   string
	err      error
}

func (req *testRequest) JobID() JobKey       { return JobKey(req.jobid) }
func (req *testRequest) WorkerID() WorkerKey { return WorkerKey(req.workerid) }

func (res *testResponse) Success() bool { return res.err == nil }
func (res *testResponse) Status() string {
	if res.err == nil {
		return "SUCCESS"
	}
	return res.err.Error()
}

func randInt(a, b int) int {
	if a == b {
		return a
	}
	if b < a {
		a, b = b, a
	}
	return a + rand.Intn(b-a)
}

func fnWork(ctx context.Context, req Request) Response {

	treq := req.(*testRequest)

	msec := randInt(treq.minMsec, treq.maxMsec)

	tres := &testResponse{
		jobid:    treq.jobid,
		workerid: treq.workerid,
		result:   fmt.Sprintf("%dms", msec),
	}

	select {
	case <-ctx.Done():
		tres.err = ctx.Err()
	case <-time.After(time.Duration(msec) * time.Millisecond):
		e := randInt(0, 100)
		if e < treq.percErr {
			tres.err = errors.New("ERR")
		}
	}

	//treq.t.Logf(" WORK:   (%s, %s) -> %s", treq.WorkerID(), treq.JobID(), tres.Status())

	return Response(tres)
}

func TestExecute(t *testing.T) {
	newreq := func(jid string, wid string) Request {
		treq := &testRequest{
			jobid:    jid,
			workerid: wid,
			minMsec:  100,
			maxMsec:  100,
			percErr:  60,
			t:        t,
		}
		return Request(treq)
	}

	workers := []*Worker{
		&Worker{"worker_1", 1, fnWork},
		&Worker{"worker_2", 1, fnWork},
		&Worker{"worker_3", 1, fnWork},
	}
	requests := []Request{
		newreq("job_1", "worker_1"),
		newreq("job_1", "worker_2"),
		newreq("job_1", "worker_3"),
		newreq("job_2", "worker_1"),
		newreq("job_2", "worker_2"),
		newreq("job_3", "worker_3"),
		newreq("job_4", "worker_1"),
		newreq("job_4", "worker_2"),
		newreq("job_5", "worker_3"),
		newreq("job_6", "worker_1"),
		newreq("job_6", "worker_2"),
		newreq("job_7", "worker_3"),
	}

	ctx := context.Background()
	out, err := Execute(ctx, workers, requests)
	if err != nil {
		t.Error(err.Error())
		return
	}
	for res := range out {
		tres := res.(*testResponse)
		t.Logf("RESULT: (%s, %s) -> %s - %s", tres.workerid, tres.jobid, tres.result, tres.Status())
	}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
