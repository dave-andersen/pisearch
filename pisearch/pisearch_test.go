package pisearch

import (
	"math/rand"
	"strconv"
	"testing"
)

var psCached *Pisearch

const (
	piFile    = "/home/dga/public_html/pi/pi200"
//pi1m"
	maxSearch = 10000000
)

// Needed to avoid duplicating openPiOrDie
type hasFatal interface {
	Fatalf(format string, args ...interface{})
}

func openPiOrDie(t hasFatal) *Pisearch {
	if psCached != nil {
		return psCached
	}
	pi, err := Open(piFile)
	if err != nil {
		t.Fatalf("Could not open Pi")
	}
	psCached = pi
	return pi
}

func TestDigitAt(t *testing.T) {
	pi := openPiOrDie(t)

	for i, wanted := range []byte{1, 4, 1, 5} {
		if d := pi.digitAt(i); d != wanted {
			t.Fatalf("digitAt(%d): %s, wanted %s", i, d, wanted)
		}
	}
}

var searchTests = []struct {
	str   string
	start int
	found bool
	pos   int
}{
	{"1", 0, true, 0},
	{"4", 0, true, 1},
	{"14", 0, true, 0},
	{"41", 0, true, 1},
	{"1415", 0, true, 0},
	{"14159", 0, true, 0},
	{"8566", 0, true, 254},
	{"85667", 0, true, 9999},
	{"856672", 0, true, 9999},
	{"8566722", 0, true, 9999},
}

func TestGetDigits(t *testing.T) {
	pi := openPiOrDie(t)

	for i, searchfor := range searchTests {
		if searchfor.found == true {
			if d := pi.GetDigits(searchfor.pos, len(searchfor.str)); d != searchfor.str {
				t.Fatalf("GetDigits(%d): %s, wanted %s", i, d, searchfor.pos)
			}
		}
	}
}

var compareTests = []struct {
	pos       int
	compareto []byte
	result    int
}{
	{0, []byte{1, 4, 1, 5}, 0},
	{0, []byte{1, 4, 1, 2}, 1},
	{0, []byte{1, 4, 1, 7}, -1},
	{1, []byte{4, 1, 5, 9}, 0},
}

func TestCompare(t *testing.T) {
	pi := openPiOrDie(t)
	for i, c := range compareTests {
		if d := pi.compare(c.pos, c.compareto); d != c.result {
			t.Fatalf("Compare(%d) pos %d vs %s: %d, wanted %d", i, c.pos, c.compareto, d, c.result)
		}
	}

}

func TestSearch(t *testing.T) {
	pi := openPiOrDie(t)
	for i, c := range searchTests {
		if f, p, _ := pi.Search(c.start, c.str); f != c.found || p != c.pos {
			t.Fatalf("Search(%d) for %s result %s %d\n", i, c.str, f, p)
		}
	}
}

func BenchmarkPisearch(b *testing.B) {
	pi := openPiOrDie(b)
	for i := 0; i < b.N; i++ {
		n := int(rand.Int31n(maxSearch))
		pi.Search(0, strconv.Itoa(n))
	}
}
