package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
)

type opts struct {
	from      string
	to        string
	offset    int64
	limit     int64
	blockSize int64
	conv      conv
}

func (o *opts) parseFlags() {
	flag.StringVar(&o.from, "from", "", "read from `FILE` instead of stdin")
	flag.StringVar(&o.to, "to", "", "write to `FILE` instead of stdout")
	flag.Int64Var(&o.offset, "offset", 0, "skip `N` bytes from the input")
	flag.Int64Var(&o.limit, "limit", 0, "read up to `N` bytes from the input")
	flag.Int64Var(&o.blockSize, "block-size", 4016, "read and write up to `BYTES` bytes at a time")
	flag.Var(&o.conv, "conv", "convert the file as per the comma separated list of `CONVS`")

	flag.Usage = usage
	flag.Parse()
}

func usage() {
	const usageMsg = "\n" +
		`Each CONV symbol may be:
	upper_case - change lower case to upper case
	lower_case - change upper case to lower case
	trim_spaces - remove all leading and trailing white spaces, as defined by Unicode`

	flag.PrintDefaults()
	fmt.Fprintln(flag.CommandLine.Output(), usageMsg)
}

func input(name string, offset, limit int64) (io.ReadCloser, error) {
	// stdin doesn't support ReadAt
	if name == "" {
		var in io.ReadCloser = os.Stdin
		if _, err := io.CopyN(io.Discard, in, offset); err != nil {
			return nil, err
		}
		if limit > 0 {
			return struct {
				io.Reader
				io.Closer
			}{io.LimitReader(in, limit), in}, nil
		}
		return in, nil
	}

	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if offset > st.Size() {
		return nil, fmt.Errorf("offset should be less tha file size")
	}
	if limit == 0 {
		// no limit
		limit = math.MaxInt64
	}
	return struct {
		io.Reader
		io.Closer
	}{io.NewSectionReader(f, offset, limit), f}, nil
}

func output(name string) (io.WriteCloser, error) {
	if name == "" {
		return os.Stdout, nil
	}
	return os.OpenFile(name, os.O_WRONLY|os.O_EXCL|os.O_CREATE, 0666)
}

func main() {
	var opts opts
	opts.parseFlags()

	if opts.limit < 0 {
		log.Fatal("limit must be non-negative")
	}
	if opts.offset < 0 {
		log.Fatal("offset must be non-negative")
	}

	in, err := input(opts.from, opts.offset, opts.limit)
	if err != nil {
		log.Fatalf("Couldn't open the input: %v", err)
	}
	defer in.Close()

	out, err := output(opts.to)
	if err != nil {
		log.Fatalf("Couldn't open the output: %v", err)
	}
	defer out.Close()

	var b bytes.Buffer
	for {
		_, err := io.CopyN(&b, in, opts.blockSize)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
	}
	out.Write(convert(b.Bytes(), opts.conv))
}
