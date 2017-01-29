package workers

import (
	"bytes"
	"fmt"
	"math/rand"
	"sort"
	"strings"
)

type mapWorkerRequests map[WorkerKey][]Request

func (wr mapWorkerRequests) String() string {

	r2s := func(r []Request) string {
		a := make([]string, 0, len(r))
		for _, x := range r {
			a = append(a, string(x.JobID()))
		}
		return "[ " + strings.Join(a, ", ") + " ]"
	}
	ws := make([]string, 0, len(wr))
	for w := range wr {
		ws = append(ws, string(w))
	}
	sort.Strings(ws)

	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "{\n")
	for _, w := range ws {
		fmt.Fprintf(buf, "   %s : %s\n", w, r2s(wr[WorkerKey(w)]))
	}
	fmt.Fprintf(buf, "}\n")
	return buf.String()
}

// ============================================================================
// RANDOMIZE
// ============================================================================

// randomize randomly permutes each worker's list of Requests
func (wr mapWorkerRequests) randomize() {
	shuffleRequests := func(src []Request) []Request {
		dest := make([]Request, len(src))
		perm := rand.Perm(len(src))
		for i, v := range perm {
			dest[v] = src[i]
		}
		return dest
	}

	for wid, items := range wr {
		wr[wid] = shuffleRequests(items)
	}
}

// ============================================================================
// DISTRIBUTE
// ============================================================================

type mapJobWorkers map[JobKey][]WorkerKey

type jobInfo struct {
	jobkey  JobKey
	workers int
}
type byLen []jobInfo

func (a byLen) Len() int           { return len(a) }
func (a byLen) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byLen) Less(i, j int) bool { return a[i].workers < a[j].workers }

// jobWorkers returns the jobWorkers map
func (wr mapWorkerRequests) jobWorkers() mapJobWorkers {
	jw := mapJobWorkers{}
	for wid, reqs := range wr {
		for _, r := range reqs {
			jid := r.JobID()
			if wids, ok := jw[jid]; ok {
				jw[jid] = append(wids, wid)
			} else {
				jw[jid] = []WorkerKey{wid}
			}
		}
	}
	//for _, w := range jw {
	//sort.Strings(w)
	//}
	return jw
}

// jobOrder returns a job key list
// ordered with the jobs with fewer requests first.
func (jw mapJobWorkers) jobOrder() []JobKey {
	// creates the jobInfo list
	var list = make([]jobInfo, 0, len(jw))
	for jid, wids := range jw {
		list = append(list, jobInfo{jid, len(wids)})
	}
	sort.Sort(byLen(list))
	a := make([]JobKey, 0, len(jw))
	for _, jobinfo := range list {
		a = append(a, jobinfo.jobkey)
	}
	return a
}

// distribute reorder each worker requests list,
// in order to have each job as soon as possible
func (src mapWorkerRequests) distribute() {
	// initialize dst
	dst := mapWorkerRequests{}
	for wid := range src {
		dst[wid] = []Request{}
	}

	for {
		jw := src.jobWorkers()
		ord := jw.jobOrder()
		if len(ord) == 0 {
			break
		}

		for _, jid := range ord {

			// list of candidate workers of the job
			wids := jw[jid]

			// select the worker with the dst shorter list of job
			var minlen, minidx int
			for idx, wid := range wids {
				l := len(dst[wid])
				if idx == 0 || l < minlen {
					minlen = l
					minidx = idx
				}
			}
			wid := wids[minidx]

			// remove the worker from the job's worker list
			wids[minidx] = wids[len(wids)-1]
			wids = wids[:len(wids)-1]
			jw[jid] = wids

			// remove the request from the src worker's request list
			// and insert into the dst worker's request  list
			reqs := src[wid]
			for idx, req := range reqs {
				if req.JobID() == jid {
					// insert the request in dst
					dst[wid] = append(dst[wid], req)
					// remove the request from src
					reqs[idx] = reqs[len(reqs)-1]
					reqs = reqs[:len(reqs)-1]
					src[wid] = reqs

					break
				}
			}

		}
	}
	for wid := range src {
		src[wid] = dst[wid]
	}
}
