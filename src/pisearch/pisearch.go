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
// performing a sequence of Search and GetDigits operations,
// and then calling Close.
//
package pisearch

import (
	"encoding/binary"
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

// Convenience function to help make Open more clear
func openAndMap(name string) (file *os.File, fi os.FileInfo, mapped []byte, err error) {
	if file, err = os.Open(name); err != nil {
		log.Println("open of", name, "failed")
		return
	}
	if fi, err = file.Stat(); err != nil {
		file.Close()
		log.Println("stat of", name, "failed")
		return
	}
	mapped, err = syscall.Mmap(int(file.Fd()), 0, int(fi.Size()), syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		file.Close()
		log.Println("mmap of ", name, "failed:", err)
	}
	return
}

// Open returns a pisearch object that references the two files
// name.4.idx and name.4.bin, or error if the files could not
// be opened and memory mapped.
func Open(name string) (*Pisearch, error) {
	file, fi, filemap, err := openAndMap(name + ".4.bin")
	if err != nil {
		return nil, err
	}

	numdigits := fi.Size() * 2

	idxfile, _, idxmap, err := openAndMap(name + ".4.idx")
	if err != nil {
		syscall.Munmap(filemap)
		file.Close()
		return nil, err
	}

	return &Pisearch{file, filemap, int(numdigits), idxfile, idxmap}, nil
}

// Close closes the pisearch object.  Note:  This code is not thread-safe.
// The caller must guarantee that no other threads are accessing the object.
func (p *Pisearch) Close() {
	// I'm writing the code this way
	// as a reminder to my future-self that, if you really want
	// to have threads playing willy-nilly, you'll need to guard
	// filemap_ and idxmap_.
	p.numDigits = 0
	tmp := p.filemap_
	p.filemap_ = nil
	_ = syscall.Munmap(tmp)
	p.pifile_.Close()
	tmp = p.idxmap_
	p.idxmap_ = nil
	_ = syscall.Munmap(tmp)
	p.idxfile_.Close()
}

// Return the digit at position p.  Requires that pos be contained
// within the file or may cause a program crash.
func (p *Pisearch) digitAt(pos int) byte {
	b := p.filemap_[pos/2]
	if (pos & 0x01) == 1 { // Second digit in a byte
		return b & 0x0f
	}
	return b >> 4
}

// GetDigits returns an ASCII string representation of the digits of
// pi from position start to min(start+length, end of pi file).
func (p *Pisearch) GetDigits(start int, length int) (digits string) {
	if start >= p.numDigits {
		return ""
	}
	end := start + length
	if end >= p.numDigits {
		end = p.numDigits - 1
	}
	outlen := end - start
	res := make([]uint8, outlen)
	for i := 0; i < outlen; i++ {
		res[i] = p.digitAt(start+i) + '0'
	}
	return string(res)
}

func (p *Pisearch) seqsearch(start int, searchkey []byte) (found bool, position int, nMatches int) {
	maxPos := p.numDigits - len(searchkey)
	for position = start; position < maxPos; position++ {
		// XXX SPEED: can optimize using the tricks in the C++ version.
		// For now, this first digitAt check avoids most of the safety
		// checks in compare at relatively low cost.
		if p.digitAt(position) == searchkey[0] {
			if p.compare(position, searchkey) == 0 {
				return true, position, 0
			}
		}
	}
	// End of Pi
	return false, 0, 0
}

/* Returns -1 if pi[start] < searchkey;
 *          0 if equal
 *          1 if >
 */

func (p *Pisearch) compare(start int, searchkey []byte) int {
	skl := len(searchkey)
	def := 0
	if (skl + start) >= p.numDigits {
		skl = p.numDigits - start
		def = -1
	}
	for i := 0; i < skl; i++ {
		da := p.digitAt(start + i)
		if da < searchkey[i] {
			return -1
		} else if da > searchkey[i] {
			return 1
		}
	}
	return def
}

func (p *Pisearch) idxAt(pos int) int {
	i := pos * 4
	return int(binary.LittleEndian.Uint32(p.idxmap_[i : i+4]))
}

func (p *Pisearch) idxsearch(start int, searchkey []byte) (found bool, position int, nMatches int) {
	i := sort.Search(p.numDigits, func(i int) bool {
		return p.compare(p.idxAt(i), searchkey) >= 0
	})
	j := i + sort.Search(p.numDigits-i, func(j int) bool {
		return p.compare(p.idxAt(j+i), searchkey) != 0
	})

	nMatches = (j - i)
	positions := make([]int, nMatches)
	for x := 0; i < j; i++ {
		positions[x] = p.idxAt(i)
		x++
	}
	if nMatches > 1 {
		sort.Ints(positions)
	}

	for _, pos := range positions {
		if pos >= start {
			return true, pos, nMatches
		}
	}
	return false, 0, 0
}

// Search returns the position at which the first instance of "searchkey"
// occurs after position "start".  Start is a zero-based offset within
// Pi (i.e., to search from the beginning, start should be zero).  If the
// key is not found, returns found=false.  This function dispatches
// to sequential and indexed search based upon the setting of seqThresh.
// nMatches will be non-zero if the index was used, or zero otherwise.
func (p *Pisearch) Search(start int, searchkey string) (found bool, position int, nMatches int) {
	querylen := len(searchkey)
	if querylen == 0 {
		return false, 0, 0
	}
	searchbytes := make([]byte, len(searchkey))
	for i := 0; i < len(searchkey); i++ {
		searchbytes[i] = searchkey[i] - '0'
	}

	if querylen <= seqThresh {
		return p.seqsearch(start, searchbytes)
	}
	return p.idxsearch(start, searchbytes)
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
