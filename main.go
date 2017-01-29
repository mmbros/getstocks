package main

import (
	"fmt"
	"sort"
	"strings"
)

//func main2() {
//os.Exit(cli.Run())
//}

type requests []string
type workerRequests map[string]requests
type jobWorkers map[string][]string

type jobInfo struct {
	job     string
	workers int
}
type byLen []jobInfo

func (a byLen) Len() int           { return len(a) }
func (a byLen) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byLen) Less(i, j int) bool { return a[i].workers < a[j].workers }

func a2s(a []string) string {
	return "[" + strings.Join(a, ", ") + "]"
}

func (wr workerRequests) Print(name string) {

	ws := make([]string, 0, len(wr))
	for w := range wr {
		ws = append(ws, w)
	}
	sort.Strings(ws)
	fmt.Printf("%s = {\n", name)
	for _, w := range ws {
		fmt.Printf("   %s : %s\n", w, a2s(wr[w]))
	}
	fmt.Printf("}\n")
}
func (jw jobWorkers) Print(name string) {

	jobs := make([]string, 0, len(jw))
	for j := range jw {
		jobs = append(jobs, j)
	}
	sort.Strings(jobs)

	fmt.Printf("%s = {\n", name)
	for _, j := range jobs {
		fmt.Printf("   %s : %s\n", j, a2s(jw[j]))
	}
	fmt.Printf("}\n")
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
		m = append(m, jobInfo{j, len(w)})
	}
	sort.Sort(byLen(m))
	a := make([]string, 0, len(jw))
	for _, i := range m {
		a = append(a, i.job)
	}
	return a

}

func main() {
	src := workerRequests{
		"w1": requests{"j1", "j2", "j4", "j6", "j9"},
		"w2": requests{"j1", "j2", "j4", "j6", "j8"},
		"w3": requests{"j1", "j3", "j5", "j7", "j8"},
	}
	dst := workerRequests{}
	for w := range src {
		dst[w] = requests{}
	}
	src.Print("src")

	for iter := 0; ; iter++ {
		fmt.Printf("====  ITER %d  ==================\n", iter+1)

		jw := src.jobWorkers()
		jw.Print("jobs")

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

		src.Print("src")
		dst.Print("dst")
	}
}
