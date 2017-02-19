package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

const (
	clLogfile = "logfile"
	// default values
	defaultLogFile = "../getstocks.log"
)

type clArgs struct {
	logfile string
}

var sessions Sessions

func parseArgs() *clArgs {
	var args clArgs
	// command line arguments
	flag.StringVar(&args.logfile, clLogfile, defaultLogFile, "Logfile to use.")
	flag.Parse()
	return &args
}

func jsonGetSessionByIndex(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	vars := mux.Vars(r)
	index, err := strconv.Atoi(vars["index"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	result := sessions.Item(index)
	if result == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := json.NewEncoder(w).Encode(result); err != nil {
		panic(err)
	}
}

func jsonGetSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	result := struct {
		Length int `json:"length"`
	}{
		Length: sessions.Length(),
	}
	if err := json.NewEncoder(w).Encode(result); err != nil {
		panic(err)
	}
}
func logger(inner http.HandlerFunc, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		log.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	}
}

func main() {
	var err error
	args := parseArgs()
	sessions, err = NewSessions(args.logfile)
	if err != nil {
		log.Fatal(err)
	}

	//for i, s := range sessions {
	//fmt.Printf("[%d] %d events, start %v, elapsed %v\n", i, len(s.events), s.start.Format("2006-01-02 15:04:05"), s.elapsed())
	//}
	r := mux.NewRouter()
	r.HandleFunc("/sessions", logger(jsonGetSessions, "jsonGetSessions")).Methods("GET")
	r.HandleFunc("/sessions/{index}", logger(jsonGetSessionByIndex, "jsonGetSessionByIndex")).Methods("GET")

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8888",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())

	//server := rpc.NewServer()
	//server.Register(sessions)
	//server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	//listener, e := net.Listen("tcp", ":8888")
	//if e != nil {
	//log.Fatal("listen error:", e)
	//}
	//for {
	//if conn, err := listener.Accept(); err != nil {
	//log.Fatal("accept error: " + err.Error())

	//} else {
	//log.Printf("new connection established: %v\n", conn.RemoteAddr())
	//go server.ServeCodec(jsonrpc.NewServerCodec(conn))
	//}
	//}
}
