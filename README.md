pisearch
========

Go language Pi search code.  Requires a file of the digits of Pi
and a suffix array built atop them:
    <name>.4.bin -- digits of pi, BCD, packed 2 digits per byte
    <name>.4.idx -- a suffix array indexing those digits, stored
                    as little-endian uint32s.

I've put a copy of pi1m.4.bin and pi1m.4.idx at:
  http://moo.cmcl.cs.cmu.edu/~dga/pi1m.4.bin
  http://moo.cmcl.cs.cmu.edu/~dga/pi1m.4.idx

if you want to play with them.  You will need some kind of file of
Pi digits in order to run the tests.

To build:
	go build cmd      (command-line search interface)
	go build piweb    (web-based/json search code)
	go build pipack   (utility to pack and unpack BCD)

	go test pisearch   (requires a pi file - see pisearch_test.go)
