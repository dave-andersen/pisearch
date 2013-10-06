package main

import (
	"encoding/json"
	"github.com/dave-andersen/pisearch/pisearch"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"
)

const (
	pifile  = "pi1m"
	LOGFILE = "/local/logs/pi/pilog"
)

var (
	logfile *os.File
)

// Return codes for JSON.  Shouldn't we use a standard, though?
const (
	STATUS_FAILED  = "FAILED"
	STATUS_SUCCESS = "success"
)

type SearchResponse struct {
	SearchKey    string `json:"k"`
	Start        int    `json:"st"`
	Status       string `json:"status"`
	Position     int    `json:"p":`
	DigitsBefore string `json:"db"`
	DigitsAfter  string `json:"da"`
}

type Piserver struct {
	searcher *pisearch.Pisearch
	logfile  *os.File
}

type jsonhandler func(*http.Request, map[string]interface{})

func (handler jsonhandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	startTime := time.Now()
	results := make(map[string]interface{})
	if err := req.ParseForm(); err != nil {
		results["status"] = STATUS_FAILED
		results["error"] = "Bad form"
	} else {
		handler(req, results)
	}

	w.Header().Set("Content-Type", "text/javascript")
	tn := time.Now()
	results["elapsedTime"] = tn.Sub(startTime)
	//b, err := json.MarshalIndent(results, "", "  ")
	b, err := json.Marshal(results)
	if err != nil {
		io.WriteString(w, "Internal error - can't marshal output\n")
		return
	}
	if b != nil {
		io.WriteString(w, string(b))
	}
	if logfile != nil {
		results["queryTime"] = tn.UnixNano()
		b, err := json.Marshal(results)
		if err == nil {
			bstr := string(b)
			logfile.WriteString(bstr + "\n")
		}
	}
}

func (ps *Piserver) ServeDigits(req *http.Request, results map[string]interface{}) {
	results["status"] = STATUS_FAILED
	startstr, has_start := req.Form["start"]
	countstr, has_count := req.Form["count"]
	if !has_start || !has_count {
		results["error"] = "Missing query parameters"
		return
	}
	start64, err := strconv.Atoi(startstr[0])
	if err != nil {
		results["error"] = "Bad start position"
		return
	}
	start := int(start64)
	count, err := strconv.Atoi(countstr[0])
	if err != nil {
		results["error"] = "Bad count"
		return
	}
	results["status"] = STATUS_SUCCESS
	results["start"] = start
	results["count"] = count
	results["digits"] = ps.searcher.GetDigits(start, count)
}

func (ps *Piserver) ServeQuery(req *http.Request, results map[string]interface{}) {
	// results["status"] = ...
	// results["results"] = [ [result1], [result2], ... ]
	results["status"] = "OK"
	q, has_q := req.Form["q"]
	if !has_q {
		results["status"] = STATUS_FAILED
		results["error"] = "Missing query"
		return
	}

	if len(q) > 20 {
		results["status"] = STATUS_FAILED
		results["error"] = "Too many queries"
		return
	}

	start_pos := 0
	start, has_start := req.Form["qs"]
	if has_start {
		sp, err := strconv.Atoi(start[0])
		if err != nil {
			results["status"] = STATUS_FAILED
			results["error"] = "Bad start position"
			return
		}
		start_pos = int(sp)
	}
	resarray := make([]SearchResponse, len(q))
	results["results"] = resarray
	for idx, query := range q {
		r := SearchResponse{SearchKey: query, Start: start_pos}
		if start_pos > 0 {
			start_pos -= 1
		}
		found, pos, _ := ps.searcher.Search(start_pos, query)
		if found {
			digitBeforeStart := pos - 20
			if digitBeforeStart < 0 {
				digitBeforeStart = 0
			}
			r.Status = "found"
			r.Position = pos + 1 // 1 based indexing for humans
			r.DigitsBefore = ps.searcher.GetDigits(digitBeforeStart, pos-digitBeforeStart)
			r.DigitsAfter = ps.searcher.GetDigits(pos+len(query), 20)
		} else {
			r.Status = "notfound"
		}
		resarray[idx] = r
	}
}

func main() {
	pisearch, err := pisearch.Open(pifile)
	if err != nil {
		log.Fatal("Could not open ", pifile, ": ", err)
	}
	logfile, err = os.OpenFile(LOGFILE, syscall.O_RDWR|syscall.O_CREAT, 0644)
	if err != nil {
		logfile = nil
	}
	server := &Piserver{pisearch, logfile}
	http.Handle("/piquery",
		jsonhandler(func(req *http.Request, respmap map[string]interface{}) {
			server.ServeQuery(req, respmap)
		}))
	http.Handle("/pidigits",
		jsonhandler(func(req *http.Request, respmap map[string]interface{}) {
			server.ServeDigits(req, respmap)
		}))

	werr := http.ListenAndServe(":1415", nil)
	if werr != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
