package pisearch

import (
	"math/rand"
	"strconv"
	"testing"
)

var psCached *Pisearch

const (
	piFile    = "pi1m"
	maxSearch = 10000000
)

func openPiOrDie(t *testing.T) *Pisearch {
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

type digitTest struct {
	pos    int
	result string
}

var digitTests []digitTest = []digitTest{
	{0, "1415"},
	{1, "4159"},
}

func TestGetDigits(t *testing.T) {
	pi := openPiOrDie(t)

	for i, searchfor := range digitTests {
		if d := pi.GetDigits(searchfor.pos, len(searchfor.result)); d != searchfor.result {
			t.Fatalf("GetDigits(%d): %s, wanted %s", i, d, searchfor.result)
		}
	}
}

type compareTest struct {
	pos       int
	compareto []byte
	result    int
}

var compareTests []compareTest = []compareTest{
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

func openPiOrDieBench(b *testing.B) *Pisearch {
	if psCached != nil {
		return psCached
	}
	pi, err := Open(piFile)
	if err != nil {
		b.Fatalf("Could not open Pi")
	}
	psCached = pi
	return pi
}

func BenchmarkPisearch(b *testing.B) {
	pi := openPiOrDieBench(b)
	for i := 0; i < b.N; i++ {
		n := int(rand.Int31n(maxSearch))
		pi.Search(0, strconv.Itoa(n))
	}
}
