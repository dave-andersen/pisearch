package main

import (
	"flag"
	"pisearch"
	//"strconv"
	"fmt"
	"log"
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Println("Too few arguments")
		return
	}
	searchstr := flag.Arg(0)

	ps, err := pisearch.Open("pi1m")
	if err != nil {
		log.Fatal("Could not open pi:", err)
	}
	found, pos := ps.Search(0, searchstr)
	fmt.Println("Found? : ", found)
	fmt.Println("Pos? : ", pos)
	ps.Close()
}
