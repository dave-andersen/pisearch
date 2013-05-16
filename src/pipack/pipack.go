package main

/* 
This program converts between 8 bit packed binary-coded decimal
format and ASCII representations of numbers.  Within each group of two
digits stored in a single byte, the leftmost is stored in the
higher-order bits of the byte.

By default it operates in pack mode.  Implementing unpack is left
as an exercise for the reader.
*/

import (
	"bufio"
	"flag"
	"os"
)

var doUnpack = flag.Bool("unpack", false, "unpack binary pi (default: pack into binary)")

func main() {
	flag.Parse()

	r := bufio.NewReader(os.Stdin)
	o := bufio.NewWriter(os.Stdout)

	b := byte(0)
	c := byte(0)

	var err error

	if !*doUnpack {
		for {
			if c, err = r.ReadByte(); err == nil {
				b = (c - '0') << 4
				if c, err = r.ReadByte(); err == nil {
					b |= (c - '0')
				}
				o.WriteByte(b)
			}
			if err != nil {
				break
			}
		}
	} else {
		for err == nil {
			if c, err = r.ReadByte(); err == nil {
				o.WriteByte((c >> 4) + '0')
				o.WriteByte((c & 0xf) + '0')
			}
		}
	}

	o.Flush()
}
