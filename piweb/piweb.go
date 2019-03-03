package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dave-andersen/pisearch/pisearch"
	"github.com/dustin/go-humanize"
	"github.com/NYTimes/gziphandler"
	"github.com/rs/cors"
)

const (
	pifile                  = "/home/dga/public_html/pi/pi200"
	LOGFILE                 = "/local/logs/pi/pilog"
	MAX_QUERIES_PER_REQUEST = 20
)

var (
	logfile    *os.File
	listenPort = flag.Int("p", 1415, "port to listen on")
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
	Position     int    `json:"p"`
	DigitsBefore string `json:"db"`
	DigitsAfter  string `json:"da"`
	Count        int    `json:"c"`
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
	results["et"] = tn.Sub(startTime)
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
		results["qt"] = tn.UnixNano()
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
	q, has_q := req.Form["q"]
	if !has_q {
		results["status"] = STATUS_FAILED
		results["error"] = "Missing query"
		return
	}

	if len(q) > MAX_QUERIES_PER_REQUEST {
		results["status"] = STATUS_FAILED
		results["error"] = "Too many queries"
		return
	}

	start_pos := 0
	if start, has_start := req.Form["qs"]; has_start {
		sp, err := strconv.Atoi(start[0])
		if err != nil {
			results["status"] = STATUS_FAILED
			results["error"] = "Bad start position"
			return
		}
		start_pos = int(sp)
	}
	resarray := make([]SearchResponse, len(q))
	results["status"] = "OK"
	results["r"] = resarray
	for idx, query := range q {
		r := SearchResponse{SearchKey: query, Start: start_pos}
		if start_pos > 0 {
			start_pos -= 1
		}
		found, pos, nMatches := ps.searcher.Search(start_pos, query)
		if found {
			digitBeforeStart := pos - 20
			if digitBeforeStart < 0 {
				digitBeforeStart = 0
			}
			r.Status = "found"
			r.Position = pos + 1 // 1 based indexing for humans
			r.Count = nMatches
			r.DigitsBefore = ps.searcher.GetDigits(digitBeforeStart, pos-digitBeforeStart)
			r.DigitsAfter = ps.searcher.GetDigits(pos+len(query), 20)
		} else {
			r.Status = "notfound"
		}
		resarray[idx] = r
	}
}

// Strings for the LegacyServer, which imitates the old CGI interface.
const (
	ERRMSG_NO_SEARCH      = "Please specify a search string\n"
	ERRMSG_INVALID_SEARCH = "You can only search for digits (0-9), try again.<br />\n" +
		"As an example, the search \"pi\" is invalid, but \n" +
		"the search \"3232\" is perfectly OK.  Don't put in\n" +
		"the quote marks, of course.  Searches can be up to 100 digits.\n"
	ERRMSG_POS_TOO_FAR = "You can't start searching for something after the last digit of pi that we have!\n"
	MAX_OUTPUT_LEN     = 1000
	MAX_SEARCH_LEN     = 100
	MY_URL             = "http://www.angio.net/pi/bigpi.cgi"
	FILE_DIR           = "/home/dga/public_html/pi"
	FOOT_DIR           = "feet"
	N_FEET             = 4
)

var (
	HeaderFileContents []byte
	FooterFileStart    []string
	FooterFileEnd      []string
)

func InitLegacy() {
	var err error
	HeaderFileContents, err = ioutil.ReadFile(path.Join(FILE_DIR, "pisearch.head.html"))
	if err != nil {
		log.Fatal("Could not read pisearch head file in legacy")
	}
	FooterFileStart = make([]string, N_FEET)
	FooterFileEnd = make([]string, N_FEET)
	for i := 0; i < N_FEET; i++ {
		contents, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/%d.html",
			FILE_DIR, FOOT_DIR, i))
		if err != nil {
			log.Fatal("Error reading footer file")
		}
		splitbits := strings.SplitN(string(contents), "PROCSEC", 2)
		FooterFileStart[i] = splitbits[0]
		FooterFileEnd[i] = splitbits[1]
	}
}

func (ps *Piserver) ServeLegacy(w http.ResponseWriter, req *http.Request) {
	startTime := time.Now()
	if err := req.ParseForm(); err != nil {
		io.WriteString(w, "Error parsing form submission: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(HeaderFileContents)

	searchkey := strings.TrimSpace(req.FormValue("UsrQuery"))
	startpos := strings.TrimSpace(req.FormValue("startpos"))
	querytype := req.FormValue("querytype")

	if len(searchkey) == 0 {
		LegacyError(w, ERRMSG_NO_SEARCH)
		return
	}
	if len(searchkey) > MAX_SEARCH_LEN {
		LegacyError(w, ERRMSG_INVALID_SEARCH)
		return
	}
	for _, c := range searchkey {
		if c < '0' || c > '9' {
			LegacyError(w, ERRMSG_INVALID_SEARCH)
			return
		}
	}

	startposd := 0
	if len(startpos) > 0 {
		var err error
		startposd, err = strconv.Atoi(startpos)
		if err != nil {
			LegacyError(w, "Invalid start position;  must be a number")
			return
		}
		// The legacy pi searcher deals in "human" 1-based indexing, unfortunately.
		if startposd > 0 {
			startposd -= 1
		}
		if startposd < 0 {
			LegacyError(w, "Can't start at a negative position")
			return
		}
		if startposd >= ps.searcher.NumDigits() {
			LegacyError(w, "Start position greater than number of digits")
			return
		}
	}

	if querytype == "substr" {
		qlen, err := strconv.Atoi(searchkey)
		if err != nil || qlen > MAX_OUTPUT_LEN ||
			(startposd+qlen) > ps.searcher.NumDigits() {
			LegacyError(w, "Invalid substring length and start")
			return
		}

		io.WriteString(w, "<div id=\"showheader\"><h3>Pi from ")
		io.WriteString(w, strconv.Itoa(startposd+1))
		io.WriteString(w, " to ")
		io.WriteString(w, strconv.Itoa(startposd+qlen))
		io.WriteString(w, "</h3></div><div id=\"showstring\"><p>")
		io.WriteString(w, ps.searcher.GetDigits(startposd, qlen))
		io.WriteString(w, "</p></div>\n")
	} else {
		// Search
		found, pos, _ := ps.searcher.Search(startposd, searchkey)
		if !found {
			io.WriteString(w, "The string ")
			io.WriteString(w, searchkey)
			io.WriteString(w, " did not occur in the first ")
			io.WriteString(w, strconv.Itoa(ps.searcher.NumDigits()))
			io.WriteString(w, " digits of pi after position ")
			io.WriteString(w, strconv.Itoa(startposd))
			io.WriteString(w, ".<br />(Sorry!  Don't give up, Pi contains lots of other cool strings.)\n")
		} else {
			pos1commas := humanize.Comma(int64(pos + 1))
			io.WriteString(w, "The string <b>")
			io.WriteString(w, searchkey)
			io.WriteString(w, "</b> occurs at position ")
			io.WriteString(w, pos1commas)
			io.WriteString(w, " counting from the first digit after the decimal point. The 3. is not counted.\n"+
				"<form method=\"post\" action=\"")
			io.WriteString(w, MY_URL)
			io.WriteString(w, "\"><input type=\"hidden\" name=\"UsrQuery\" value=\"")
			io.WriteString(w, searchkey)
			io.WriteString(w, "\"><input type=\"hidden\" name=\"startpos\" value=\"")
			io.WriteString(w, strconv.Itoa(pos+2))
			io.WriteString(w, "\"><input type=\"submit\" value=\"Find Next\">"+
				"</form>\n"+
				"<p>The string and surrounding digits:</p><p>\n")
			if pos <= 20 {
				io.WriteString(w, ps.searcher.GetDigits(0, 20))
			} else {
				io.WriteString(w, ps.searcher.GetDigits(pos-20, 20))
			}
			io.WriteString(w, "<b>")
			io.WriteString(w, ps.searcher.GetDigits(pos, len(searchkey)))
			io.WriteString(w, "</b>")
			io.WriteString(w, ps.searcher.GetDigits(pos+len(searchkey), 20))
			io.WriteString(w, "</p>")
		}
	}

	endTime := time.Now()
	elapsed := endTime.Sub(startTime)
	PrintFooter(w, elapsed.String())
}

func LegacyError(w http.ResponseWriter, errmsg string) {
	io.WriteString(w, "<div class=\"errmsg\">")
	io.WriteString(w, errmsg)
	io.WriteString(w, "</div>")
	PrintFooter(w, "")
}

func PrintFooter(w http.ResponseWriter, procsec string) {
	// Template substitute procsec into the footer template
	footerFile := rand.Int() % N_FEET
	io.WriteString(w, FooterFileStart[footerFile])
	io.WriteString(w, procsec)
	io.WriteString(w, FooterFileEnd[footerFile])
}

func main() {
	flag.Parse()
	InitLegacy()

	pisearch, err := pisearch.Open(pifile)
	if err != nil {
		log.Fatal("Could not open ", pifile, ": ", err)
	}
	logfile, err = os.OpenFile(LOGFILE, syscall.O_RDWR|syscall.O_CREAT, 0644)
	if err != nil {
		logfile = nil
	}
	server := &Piserver{pisearch, logfile}
	mux := http.NewServeMux()
	mux.Handle("/piquery",
		jsonhandler(func(req *http.Request, respmap map[string]interface{}) {
			server.ServeQuery(req, respmap)
		}))
	mux.Handle("/pidigits",
		jsonhandler(func(req *http.Request, respmap map[string]interface{}) {
			server.ServeDigits(req, respmap)
		}))
	handleBigPi := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			server.ServeLegacy(w, r)
		})
	mux.Handle("/bigpi.cgi", gziphandler.GzipHandler(handleBigPi))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://angio.net", "http://www.angio.net",
			"https://angio.net", "https://www.angio.net"},
	})
	handler := c.Handler(mux)

	listenPortString := fmt.Sprintf(":%d", *listenPort)
	werr := http.ListenAndServe(listenPortString, handler)
	if werr != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
