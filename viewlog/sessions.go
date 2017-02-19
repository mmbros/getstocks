package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

type event struct {
	Level       string    `json:"level"`
	Msg         string    `json:"msg"`
	Time        time.Time `json:"time"`
	TimeStart   time.Time `json:"time_start"`
	TimeEnd     time.Time `json:"time_end"`
	Scraper     string    `json:"scraper"`
	ScraperInst int       `json:"scraper_inst"`
	Stock       string    `json:"stock"`
	StockDate   string    `json:"stock_date,omitempty"`
	StockPrice  string    `json:"stock_price,omitempty"`
}

type Session struct {
	Start  time.Time `json:"start"`
	Finish time.Time `json:"finish"`
	Events []*event  `json:"events"`
}
type Sessions []*Session

func (e *event) elapsed() time.Duration {
	return e.TimeEnd.Sub(e.TimeStart)
}

func (s *Session) elapsed() time.Duration {
	return s.Finish.Sub(s.Start)
}

// NewSessions load the sessions from the file path.
func NewSessions(path string) (Sessions, error) {
	// open file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	sessions := make(Sessions, 0, 5)
	var ses *Session

	dec := json.NewDecoder(file)
	for {
		var e event
		if err := dec.Decode(&e); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if e.Msg == "PROGRAM START" {
			ses = &Session{
				Events: []*event{},
			}
			sessions = append(sessions, ses)
			ses.Start = e.Time
			continue
		}
		if ses == nil {
			return nil, fmt.Errorf("Invalid file: missing PROGRAM START line")
		}
		if e.Msg == "PROGRAM FINISH" {
			ses.Finish = e.Time
		} else {
			ses.Events = append(ses.Events, &e)
		}
	}
	return sessions, nil
}

// Length returns the number of items of the array of sessions.
func (s Sessions) Length() int {
	return len(s)
}

// Item returns the n-th session of the list.
func (s Sessions) Item(n int) *Session {
	if n < 0 || n >= len(s) {
		return nil
	}
	return s[n]
}
