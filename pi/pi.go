package main

import (
	"flag"
	"fmt"
	"github.com/dave-andersen/pisearch/pisearch"
	"log"
	"strconv"
)

const (
	MODE_SEARCH = iota
	MODE_ANALYZE
)
const (
	PIFILE_DEFAULT = "pi1m"
)

var piFile = flag.String("pifile", PIFILE_DEFAULT, "what pi base file to use")

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Println("use:  pi <command> [command args]")
		return
	}

	p, err := pisearch.Open(*piFile)
	if err != nil {
		log.Fatal("Could not open", *piFile, ":", err)
	}

	switch flag.Arg(0) {
	case "search":
		do_search(p)
	case "analyze":
		do_analyze(p)
	case "count":
		do_count(p)
	}
	p.Close()
}

func do_search(p *pisearch.Pisearch) {
	if flag.NArg() < 2 {
		log.Fatal("use: pi search <string> [startpos]")
	}
	searchstr := flag.Arg(1)
	startpos := 0
	if flag.NArg() > 2 {
		startpos, _ = strconv.Atoi(flag.Arg(1))
	}
	found, pos, nMatches := p.Search(startpos, searchstr)
	fmt.Println("Found:  ", found)
	fmt.Println("Pos:    ", pos)
	fmt.Println("nMatch: ", nMatches)
}

func do_count(p *pisearch.Pisearch) {
	if flag.NArg() < 2 {
		log.Fatal("use: pi count <string>")
	}
	searchstr := flag.Arg(1)
	fmt.Println(p.Count(searchstr))
}

func do_analyze(p *pisearch.Pisearch) {
	fmt.Println("This function is not yet implemented")
	return
	for i := 0; i < 1000; i++ {
		str := fmt.Sprintf("%3.3d", i)
		fmt.Println(str)
	}
}
