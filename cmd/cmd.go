package main

import (
	"flag"
	"github.com/dave-andersen/pisearch/pisearch"
	"strconv"
	"fmt"
	"log"
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Println("use:  cmd <search string>  [start position]")
		return
	}
	searchstr := flag.Arg(0)
	startpos := 0
	if flag.NArg() > 1 {
		startpos, _ = strconv.Atoi(flag.Arg(1))
	}
	ps, err := pisearch.Open("pi1m")
	if err != nil {
		log.Fatal("Could not open pi:", err)
	}
	found, pos, _ := ps.Search(startpos, searchstr)
	fmt.Println("Found? : ", found)
	fmt.Println("Pos? : ", pos)
	ps.Close()
}
