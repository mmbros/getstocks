package workers

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"testing"
)

type requests []string
type workerRequests map[string]requests
type jobWorkers map[string][]string

func a2s(a []string) string {
	return "[" + strings.Join(a, ", ") + "]"
}

func (wr workerRequests) String() string {

	ws := make([]string, 0, len(wr))
	for w := range wr {
		ws = append(ws, w)
	}
	sort.Strings(ws)

	buf := &bytes.Buffer{}
	fmt.Fprintln(buf, "{")
	for _, w := range ws {
		fmt.Fprintf(buf, "   %s : %s\n", w, a2s(wr[w]))
	}
	fmt.Fprintln(buf, "}")
	return buf.String()
}
func (jw jobWorkers) String() string {

	jobs := make([]string, 0, len(jw))
	for j := range jw {
		jobs = append(jobs, j)
	}
	sort.Strings(jobs)

	buf := &bytes.Buffer{}
	fmt.Fprintln(buf, "{")
	for _, j := range jobs {
		fmt.Fprintf(buf, "   %s : %s\n", j, a2s(jw[j]))
	}
	fmt.Fprintln(buf, "}")
	return buf.String()
}

// return a map: job -> number of request with the job
func (wr workerRequests) jobWorkers() jobWorkers {
	m := jobWorkers{}
	for w, r := range wr {
		for _, j := range r {
			if ws, ok := m[j]; ok {
				m[j] = append(ws, w)
			} else {
				m[j] = []string{w}
			}
		}
	}
	for _, w := range m {

		sort.Strings(w)
	}
	return m
}

func (jw jobWorkers) order() []string {
	var m = make([]jobInfo, 0, len(jw))
	for j, w := range jw {
		m = append(m, jobInfo{JobKey(j), len(w)})
	}
	sort.Sort(byLen(m))
	a := make([]string, 0, len(jw))
	for _, i := range m {
		a = append(a, string(i.jobkey))
	}
	return a

}

func TestMain(t *testing.T) {

	src := workerRequests{
		"w1": requests{"j1", "j2", "j4", "j6", "j9"},
		"w2": requests{"j1", "j2", "j4", "j6", "j8"},
		"w3": requests{"j1", "j3", "j5", "j7", "j8"},
	}
	dst := workerRequests{}
	for w := range src {
		dst[w] = requests{}
	}
	t.Logf("src = %s", src)

	for iter := 0; ; iter++ {
		fmt.Printf("====  ITER %d  ==================\n", iter+1)

		jw := src.jobWorkers()
		t.Logf("jobs = %s\n", jw)

		ord := jw.order()
		fmt.Printf("ord = %s\n", a2s(ord))
		if len(ord) == 0 {
			break
		}

		for _, j := range ord {

			// list of candidate workers of the job
			ws := jw[j]

			// select the worker with the dst shorter queue of job
			var minlen, minidx int
			for i, w := range ws {
				l := len(dst[w])
				if i == 0 || l < minlen {
					minlen = l
					minidx = i
				}
			}
			w := ws[minidx]
			// remove the worker from the job's worker list
			ws[minidx] = ws[len(ws)-1]
			ws = ws[:len(ws)-1]
			jw[j] = ws

			// remove the job from the src worker's job queue
			a := src[w]
			for i, jj := range a {
				if jj == j {
					a[i] = a[len(a)-1]
					a = a[:len(a)-1]
					src[w] = a
					break
				}
			}

			// insert the job in the dest worker's job queue
			dst[w] = append(dst[w], j)
		}

		t.Logf("src = %s", src)
		t.Logf("dst = %s", dst)
	}
}
