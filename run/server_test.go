package run

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func getMsec(smin, smax string, def int) int {
	var a, b int
	var aerr, berr error

	a, aerr = strconv.Atoi(smin)
	b, berr = strconv.Atoi(smax)

	if aerr != nil {
		if berr == nil {
			a = b
		} else {
			a = def
		}
	}
	if berr != nil {
		b = a
	}
	if b < a {
		a, b = b, a
	}
	if a < 0 {
		a = 0
		if b < a {
			b = a
		}
	}

	msec := a
	if a != b {
		msec += rand.Intn(b - a)
	}
	return msec
}

func TestGetMsec(t *testing.T) {
	const def = 100
	var testCases = []struct {
		smin, smax string
		min, max   int
	}{
		{"", "", def, def},
		{"10", "2000", 10, 2000},
		{"", "1000", 1000, 1000},
		{"1000", "", 1000, 1000},
		{"-10", "-10", 0, 0},
	}
	for _, tc := range testCases {
		msec := getMsec(tc.smin, tc.smax, def)
		if msec < tc.min || msec > tc.max {
			t.Errorf("(smin, smax)=(%q,%q): got %d, expected in (%d, %d)", tc.smin, tc.smax, msec, tc.min, tc.max)
		}

	}
}

func returnError(sThreshold string) bool {
	threshold, _ := strconv.Atoi(sThreshold)
	if threshold < 0 {
		threshold = 0
	} else if threshold > 100 {
		threshold = 100
	}
	n := rand.Intn(101) // range 0 .. 10
	return n < threshold
}

func TestReturnError(t *testing.T) {
	const (
		iter      = 1000
		threshold = 30
	)
	rand.Seed(time.Now().UTC().UnixNano())

	s := strconv.Itoa(threshold)
	count := 0

	for j := 0; j < iter; j++ {
		if returnError(s) {
			count++
		}
	}

	expect := (threshold * iter) / 100
	diff := count - expect
	if diff < 0 {
		diff = -diff
	}
	if (diff*100)/iter > 5 {

		t.Errorf("Expecting %d errors , found %d errors on %d iters", expect, count, iter)
	}

	t.Logf("Expecting %d errors , found %d errors on %d iters (diff=%d)", expect, count, iter, diff)
}

func handlerTestServer(w http.ResponseWriter, r *http.Request) {
	// msec1  = nvl(msec2, 100) if missing
	// msec2  = msec1 if missing
	// price  = msec sleeped
	// date   = now if missing
	// err    = prob of error: 0 or missing = no error, 100

	v := r.URL.Query()

	msec := getMsec(v.Get("msec1"), v.Get("msec2"), 100)

	time.Sleep(time.Duration(msec) * time.Millisecond)

	if returnError(v.Get("err")) {
		//log.Println("SERVER ERROR - " + r.URL.Path)
		status := http.StatusNotFound
		http.Error(w, http.StatusText(status), status)
		return
	}

	price := msec
	date := time.Now()
	fmt.Fprintf(w, "<ul>\n  <li>Price: <b>%d</b></li>\n  <li>Date: <b>%s</b></li>\n</ul>", price, date)
}
