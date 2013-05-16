// Copyright 2013 David G. Andersen.  All rights reserved.
// Use of this source code is goverened by a BSD-style
// license that can be found in the Go source code distribution
// LICENSE file.

// Package pisearch provides an interface to read and search
// a BCD-encoded file of the digits of Pi together with a
// suffix array index (generated separately) for those digits.
// It takes as an argument the base name of the Pi files,
// which should be named "basename.4.bin" and "basename.4.idx"
// for the BCD digits and the suffix array index, respectively.
//
// Using this code typically operates by calling Open,
// performing a sequence of Search and GetGidigits operations,
// and then calling Close.
//
package pisearch

import (
	"log"
	"os"
	"sort"
	"syscall"
)

const (
	seqThresh = 6 // Search strings >= seqThresh digits long use the index.
)

type Pisearch struct {
	pifile_   *os.File
	filemap_  []byte
	numDigits int
	idxfile_  *os.File
	idxmap_   []byte
}

// Open returns a pisearch object that references the two files
// name.4.idx and name.4.bin, or error if the files could not
// be opened and memory mapped.
func Open(name string) (pisearch *Pisearch, err error) {
	file, err := os.Open(name + ".4.bin")
	if err != nil {
		log.Println("open of .4.bin failed")
		return nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		log.Println("stat failed")
		file.Close()
		return nil, err
	}

	numdigits := fi.Size() * 2

	idxfile, err := os.Open(name + ".4.idx")
	if err != nil {
		log.Println("open of .4.idx failed")
		file.Close()
		return nil, err
	}
	idxfi, err := idxfile.Stat()
	if err != nil {
		log.Println("stat of idx failed")
		idxfile.Close()
		file.Close()
		return nil, err
	}

	filemap, err := syscall.Mmap(int(file.Fd()), 0, int(fi.Size()), syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		log.Println("mmap of file failed")
		file.Close()
		idxfile.Close()
		return nil, err
	}

	idxmap, err := syscall.Mmap(int(idxfile.Fd()), 0, int(idxfi.Size()), syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		log.Println("mmap of idx file failed")
		syscall.Munmap(filemap)
		file.Close()
		idxfile.Close()
		return nil, err
	}

	return &Pisearch{file, filemap, int(numdigits), idxfile, idxmap}, nil
}

// Close closes the pisearch object.  Note:  This code is not thread-safe.
// The caller must guarantee that no other threads are accessing the object.
func (pisearch *Pisearch) Close() {
	// I'm writing the code this way
	// as a reminder to my future-self that, if you really want
	// to have threads playing willy-nilly, you'll need to guard
	// filemap_ and idxmap_.
	pisearch.numDigits = 0
	tmp := pisearch.filemap_
	pisearch.filemap_ = nil
	_ = syscall.Munmap(tmp)
	pisearch.pifile_.Close()
	tmp = pisearch.idxmap_
	pisearch.idxmap_ = nil
	_ = syscall.Munmap(tmp)
	pisearch.idxfile_.Close()
}

// Return the digit at position p.  Requires that pos be contained
// within the file or may cause a program crash.
func (pisearch *Pisearch) digitAt(pos int) byte {
	b := pisearch.filemap_[pos/2]
	if (pos & 0x01) == 1 { // Second digit in a byte
		return b & 0x0f
	} else {
		return b >> 4
	}
}

// GetDigits returns an ASCII string representation of the digits of
// pi from position start to max(start+length, end of pi file).
func (pisearch *Pisearch) GetDigits(start int, length int) (digits string) {
	end := start + length
	if start >= pisearch.numDigits {
		return ""
	}
	if end >= pisearch.numDigits {
		end = pisearch.numDigits - 1
	}
	outlen := end - start
	res := make([]uint8, outlen)
	// XXX SPEED:  This can be optimized in bulk instead of using digitAt
	for i := 0; i < outlen; i++ {
		res[i] = pisearch.digitAt(start+i) + '0'
	}
	return string(res)
}

func (pisearch *Pisearch) seqsearch(start int, searchkey []byte) (found bool, position int) {
	maxPos := pisearch.numDigits - len(searchkey)
	for position = start; position < maxPos; position++ {
		// XXX SPEED: can optimize using the tricks in the C++ version.
		// For now, this first digitAt check avoids most of the safety
		// checks in compare at relatively low cost.
		if pisearch.digitAt(position) == searchkey[0] {
			if pisearch.compare(position, searchkey) == 0 {
				return true, position
			}
		}
	}
	// End of Pi
	return false, 0
}

/* Returns -1 if pi[start] < searchkey;
 *          0 if equal
 *          1 if >
 */

func (pisearch *Pisearch) compare(start int, searchkey []byte) int {
	skl := len(searchkey)
	def := 0
	if (skl + start) >= pisearch.numDigits {
		skl = pisearch.numDigits - start
		def = -1
	}
	for i := 0; i < skl; i++ {
		da := pisearch.digitAt(start + i)
		if da < searchkey[i] {
			return -1
		} else if da > searchkey[i] {
			return 1
		}
	}
	return def
}

func (pisearch *Pisearch) idxAt(pos int) int {
	i := pos * 4
	return int(pisearch.idxmap_[i]) | (int(pisearch.idxmap_[i+1]) << 8) | (int(pisearch.idxmap_[i+2]) << 16) | (int(pisearch.idxmap_[i+3]) << 24)
}

func (pisearch *Pisearch) idxsearch(start int, searchkey []byte) (found bool, position int) {
	i := sort.Search(pisearch.numDigits, func(i int) bool {
		return pisearch.compare(pisearch.idxAt(i), searchkey) >= 0
	})
	j := i + sort.Search(pisearch.numDigits-i, func(j int) bool {
		return pisearch.compare(pisearch.idxAt(j+i), searchkey) != 0
	})
	//fmt.Println("Compare got i: ", i, "j", j)
	//fmt.Println("Digits there: ", pisearch.GetDigits(pisearch.idxAt(i), len(searchkey)))

	nMatches := (j - i)
	var positions []int
	for ; i < j; i++ {
		positions = append(positions, pisearch.idxAt(i))
	}
	if nMatches > 1 {
		sort.Ints(positions)
	}

	for i := 0; i < nMatches; i++ {
		if positions[i] >= start {
			return true, positions[i]
		}
	}
	return false, 0
}

// Search returns the position at which the first instance of "searchkey"
// occurs after position "start".  Start is a zero-based offset within
// Pi (i.e., to search from the beginning, start should be zero).  If the
// key is not found, returns found=false.  This function dispatches
// to sequential and indexed search based upon the setting of seqThresh.
func (pisearch *Pisearch) Search(start int, searchkey string) (found bool, position int) {
	querylen := len(searchkey)
	if querylen == 0 {
		return false, 0
	}
	searchbytes := make([]byte, len(searchkey))
	for i := 0; i < len(searchkey); i++ {
		searchbytes[i] = searchkey[i] - '0'
	}

	if querylen <= seqThresh {
		return pisearch.seqsearch(start, searchbytes)
	} else {
		return pisearch.idxsearch(start, searchbytes)
	}
}

// Summary of speed improvements not taken from the C++ version:
// Optimized binary search that takes advantage of the uniform
// distribution of numbers (1/2 as many comparisons)
// two-digit-at-a-time comparisons;
// unrolled first few digit comparisons beyond first digit;
// mapping the index directly as uint32s (pins endian-ness);
// not invoking a full sort for finding the match;
//
// From benchmarking, it's likely that the most profitable optimizations
// are distribution-based search and eliminating the call to integer sort,
// if we ever care. :-)
