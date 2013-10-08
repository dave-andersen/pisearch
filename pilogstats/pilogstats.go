package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// {"elapsedTime":7014358,"queryTime":1381102625012418396,"results":[{"k":"593211","st":0,"status":"found","p":764903,"db":"10301168491367031251","da":"22314328916323155148"}],"status":"OK"}

type LogEntry struct {
	ElapsedTime int64                    `json:"elapsedTime"`
	QueryTime   int64                    `json:"queryTime"`
	Results     []map[string]interface{} `json:"results"`
}

func logHandle(le LogEntry) {
	// hardcoded - looking at distribution of time it takes to process
	// queries.  Are particular queries causing us to slow down?
	fmt.Println(le.ElapsedTime)
}

func main() {
	var logEnt LogEntry
	f, err := os.Open("/local/logs/pi/pilog")
	if err != nil {
		log.Fatal("could not open pilog", err)
	}
	defer f.Close()
	br := bufio.NewReader(f)
	for line, err := br.ReadBytes('\n'); err == nil; line, err = br.ReadBytes('\n') {
		decodeErr := json.Unmarshal(line, &logEnt)
		if decodeErr == nil {
			logHandle(logEnt)
		} else {
			fmt.Println("Unmarshal err", decodeErr)
		}
	}
}
