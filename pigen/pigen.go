package main

// This is a Go translation of Nick Craig-Wood's Python implementation
// of fixedpoint Pi computation using the basic Chudnovsky algorithm.
// There are faster ways to compute Pi - grab the GMP Pi demo program
// if you're serious about it.
// For a better explanation of the algorithm and cleaner code,
// see the original:  http://www.craig-wood.com/nick/articles/pi-chudnovsky/

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
	
	k := big.NewInt(1)
	a_k := big.NewInt(0).Set(onebase)
	a_sum := big.NewInt(0).Set(onebase)
	b_sum := big.NewInt(0)
	tmp := big.NewInt(0)

	zero, one, five, six := big.NewInt(0), big.NewInt(1), big.NewInt(5), big.NewInt(6)
	sixk := big.NewInt(0)
	t1, t2, t3, kCubed := &big.Int{}, &big.Int{}, &big.Int{}, &big.Int{}

	for {
		// -(6*k-5)*(2*k-1)*(6*k-1)
		sixk.Mul(k, six)
		t1.Sub(sixk, five)
		t2.Mul(k, big.NewInt(2)) // Lsh
		t2.Sub(t2, one)
		t3.Sub(sixk, one)
		t1.Mul(t1, t2)
		t1.Mul(t1, t3)
		a_k.Mul(a_k, t1.Neg(t1))
		kCubed.Exp(k, big.NewInt(3), nil)
		// division: preserve precision
		a_k.Div(a_k, kCubed.Mul(kCubed, C3Div24))

		a_sum.Add(a_sum, a_k)
		b_sum.Add(b_sum, tmp.Mul(a_k, k))
		
		k.Add(k, one)
		if a_k.Cmp(zero) == 0 { // translate
			break
		}
	}
	total := big.NewInt(13591409)
	tmp.SetInt64(545140134)
	total.Mul(total, a_sum)
	tmp.Mul(tmp, b_sum)
	total.Add(total, tmp)

	tmp.SetInt64(10005)
	tmp.Mul(tmp, onebase)
	tmp.Mul(tmp, onebase) // tmp = 10005*onebase^2
	// so sqrt(tmp) = sqrt(10005)*onebase with precision maintained
	tmp = mathutil.SqrtBig(tmp)
	tmp.Mul(tmp, big.NewInt(426880))
	//fmt.Println(tmp)
	//fmt.Println(total)
	tmp.Mul(tmp, onebase)
	tmp.Div(tmp, total)
	return tmp
}