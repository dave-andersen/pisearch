// pigen generates digits of Pi using
// a Go translation of Nick Craig-Wood's Python implementation
// of fixedpoint Pi computation using the binary split Chudnovsky algorithm.
// There are faster ways to compute Pi - grab the GMP Pi demo program
// if you're serious about it.
// For a better explanation of the algorithm and cleaner code,
// see the original:  http://www.craig-wood.com/nick/articles/pi-chudnovsky/
//
// I've validated the output of this up to 1 million digits.  YMMV.
// Takes 130 seconds to produce 1m digits on a 2012 Macbook Pro using
// go1.1.
//
package main

import (
	"flag"
	"fmt"
	"math/big"
	"strconv"
	"github.com/cznic/mathutil" // gives us sqrt
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Println("usage:  pigen <digits>")
		return
	}
	n, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		fmt.Println("digits must be an integer count")
	}
	n += 1 // We skip the 3...
	// Compute a few more digits - the trailing end is wrong otherwise
	// at low precision counts
	pi := chudPi(int64(n+5))
	pistr := pi.String()[1:n] // skip the 3.
	fmt.Println(pistr)
}

func chudPi(n int64) *big.Int {
	onebase := big.NewInt(10)
	onebase.Exp(onebase, big.NewInt(n), nil)
	C := big.NewInt(640320)
	C3Div24 := big.NewInt(0).Exp(C, big.NewInt(3), nil)
	C3Div24.Div(C3Div24, big.NewInt(24))
	zero, one, three, five, six := big.NewInt(0), big.NewInt(1), big.NewInt(3), big.NewInt(5), big.NewInt(6)

	var bs func(a, b *big.Int) (*big.Int, *big.Int, *big.Int)

	bs = func(a, b *big.Int) (*big.Int, *big.Int, *big.Int) {
		//fmt.Println("bs ", a, b)
		Pab, Qab, Tab := big.NewInt(1), big.NewInt(1), big.NewInt(0)
		bminusa := big.NewInt(0).Sub(b, a)
		if (one.Cmp(bminusa) == 0) {
			if (zero.Cmp(a) != 0) {
				sixa := big.NewInt(0).Mul(a, six)
				Pab.Sub(sixa, one)
				sixa.Sub(sixa, five)
				Pab.Mul(Pab, sixa) // *= 6*a-5
				sixa.Lsh(a, 1)
				sixa.Sub(sixa, one)
				Pab.Mul(Pab, sixa)

				Qab.Exp(a, three, nil)
				Qab.Mul(Qab, C3Div24)
			}
			Tab.SetInt64(545140134)
			Tab.Mul(Tab, a)
			Tab.Add(Tab, big.NewInt(13591409))
			Tab.Mul(Tab, Pab)
			if a.Bit(0) == 1 {
				Tab.Neg(Tab)
			}
		} else {
			// m is the midpoint between a & b
			m := big.NewInt(0).Add(a, b)
			m.Rsh(m, 1)
			Pam, Qam, Tam := bs(a, m)
			Pmb, Qmb, Tmb := bs(m, b)
			Pab.Mul(Pam, Pmb)
			Qab.Mul(Qam, Qmb)

			// Tab = Qmb * Tam + Pam * Tmb
			Qmb.Mul(Qmb, Tam)
			Pam.Mul(Pam, Tmb)
			Tab.Add(Qmb, Pam)
		}
		return Pab, Qab, Tab
	}
	DigitsPerTerm := int64(14)
	startB := big.NewInt(n/DigitsPerTerm + 1)

	_, q, t := bs(big.NewInt(0), startB)
	
	tmp := big.NewInt(10005)
	tmp.Mul(tmp, onebase)
	tmp.Mul(tmp, onebase) // tmp = 10005*onebase^2
	// so sqrt(tmp) = sqrt(10005)*onebase with precision maintained
	tmp = mathutil.SqrtBig(tmp)
	tmp.Mul(tmp, big.NewInt(426880))
	tmp.Mul(tmp, q)
	tmp.Div(tmp, t)
	return tmp
}
