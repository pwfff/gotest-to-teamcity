// Copyright 2016 Apcera Inc. All rights reserved.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestOutput(t *testing.T) {
	cases := []struct {
		inputFile string // the file containing the output of `go test -v`
		wantFile  string // aka the "golden" file containing the expected output
	}{
		{
			"testdata/apcera-setup.txt",
			"testdata/apcera-setup-want.txt",
		},
		{
			"testdata/cobra.txt",
			"testdata/cobra-want.txt",
		},
		{
			"testdata/docker-compat.txt",
			"testdata/docker-compat-want.txt",
		},
	}

	for _, c := range cases {
		// Open the file containing the output of `go test -v`.
		b, err := ioutil.ReadFile(c.inputFile)
		if err != nil {
			t.Fatalf("failed to open input file %q: %v", c.inputFile, err)
		}
		in = bytes.NewReader(b) // mock stdin

		// Open the file containing the expected output.
		want, err := ioutil.ReadFile(c.wantFile)
		if err != nil {
			t.Fatalf("failed to open golden file %q: %v", c.wantFile, err)
		}
		want = bytes.TrimSpace(want)

		buf := &bytes.Buffer{}
		const arbitrarilyLargeBufSize = 65535 // saves allocations
		buf.Grow(arbitrarilyLargeBufSize)
		out = buf // capture the output of main()

		main()

		got := bytes.TrimSpace(buf.Bytes())
		err = compare(got, want)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// compare does a byte-by-byte comparison between two slices. It returns an
// error if there is a difference. The error contains the first unequal bytes
// and some of the following characters for context.
func compare(got, want []byte) error {
	for i := 0; i < min(len(got), len(want)); i++ {
		if got[i] != want[i] {
			// Print the rest of the line for context.
			gotEOL := eol(got[i:])
			wantEOL := eol(want[i:])
			end := i + min(gotEOL, wantEOL)
			return fmt.Errorf("\ngot:  %s\nwant: %s", got[i:end], want[i:end])
		}
	}
	return nil
}

// eol reads the given slice up to the next newline or EOF character. It returns
// the index of the newline, or the length of b if none is found.
func eol(b []byte) int {
	for i := 0; i < len(b); i++ {
		if b[i] == '\n' {
			return i
		}
	}
	return len(b)
}

// min returns the minimum of a and b.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
