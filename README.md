pisearch
========

Go language Pi search code.  Requires a file of the digits of Pi
and a suffix array built atop them:

    <name>.4.bin -- digits of pi, BCD, packed 2 digits per byte

    <name>.4.idx -- a suffix array indexing those digits, stored
                    as little-endian uint32s.

I've put a copy of pi1m.4.bin and pi1m.4.idx at:
  - http://moo.cmcl.cs.cmu.edu/~dga/pi1m.4.bin
  - http://moo.cmcl.cs.cmu.edu/~dga/pi1m.4.idx

if you want to play with them.  You will need some kind of file of
Pi digits in `src/pisearch` in order to run the tests.

The search code is in `src/pisearch`.  It's had more thought put into
it than the quick hack of a web interface in `src/piweb`.  Use at your own risk.  The json exported by piweb is still changing a lot as I work on the javascript/html front-end.

# Building #

	cd pisearch
        export GOPATH=`/bin/pwd`
	go install pi       (command-line search interface)
	go install piweb    (web-based/json search code)
	go install pipack   (utility to pack and unpack BCD)

	go test pisearch   (requires a pi file - see pisearch_test.go)

Optionally:

	go install pigen    (generate pi using the Chudnovsky algorithm)

pigen depends on an external package that you must install using go get:

	go get github.com/cznic/mathutil

Of course, you can also check out the pi search code into a subdirectory
of your GOPATH, adjusting things as needed for your setup.  To verify that
you've got things running, try some commands such as:

	pi search 1415
	pi count 99999

