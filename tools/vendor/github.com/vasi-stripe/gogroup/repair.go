package gogroup

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"sort"

	"golang.org/x/tools/imports"
)

// Read lines from an io.Reader.
func readLines(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	ret := []string{}
	for scanner.Scan() {
		ret = append(ret, scanner.Text())
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return ret, nil
}

// Write some lines to an io.Writer.
func writeLines(w io.Writer, lines []string) error {
	for _, line := range lines {
		_, err := fmt.Fprintln(w, line)
		if err != nil {
			return err
		}
	}
	return nil
}

// Generate what the import section of a file should look like, properly
// sorted.
// Input is a set of grouped imports, and all the lines of text in the file.
// Output is the lines of text that make up the sorted import section.
func sortedImportLines(gs groupedImports, lines []string) []string {
	sort.Sort(gs)

	ret := []string{}
	var prev *groupedImport
	for _, g := range gs {
		if prev != nil && g.group != prev.group {
			// Time for an empty line.
			ret = append(ret, "")
		}
		ret = append(ret, lines[g.startLine:g.endLine+1]...)
		prev = g
	}

	return ret
}

// Given the contents of a source file and the parsed imports, yield
// the contents of the file with imports sorted and grouped, as an
// io.Reader.
func fixImports(src []byte, gs groupedImports) (io.Reader, error) {
	lines, err := readLines(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}

	min := gs[0].startLine
	max := gs[len(gs)-1].endLine

	// Need to start a new slice, or we may modify lines as we append.
	out := []string{}
	out = append(out, lines[:min]...)
	out = append(out, sortedImportLines(gs, lines)...)
	out = append(out, lines[max+1:]...)

	var dst bytes.Buffer
	if err = writeLines(&dst, out); err != nil {
		return nil, err
	}

	return &dst, nil
}

// Repair the imports section of a file, to reflect sorting and grouping.
func (p *Processor) repair(fileName string, r io.Reader) (io.Reader, error) {
	// Get the full contents.
	src, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Check if the file needs any fixing.
	gs, err := p.readImports(fileName, bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	if gs.validate() == nil {
		return nil, nil
	}

	// Generate the fixed version.
	dst, err := fixImports(src, gs)
	if err != nil {
		return nil, err
	}

	return dst, nil
}

// Both reformat the file and fix the imports section.
func (p *Processor) reformat(fileName string, r io.Reader) (io.Reader, error) {
	// Get the full contents.
	src, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	formatted, err := imports.Process(fileName, src, nil)
	if err != nil {
		return nil, err
	}

	ret, err := p.repair(fileName, bytes.NewReader(formatted))
	if err != nil {
		return nil, err
	}
	if ret == nil {
		if bytes.Equal(src, formatted) {
			// No change by either goimports or grouping.
			return nil, nil
		}

		// Format changed, but no imports rewrites needed.
		return bytes.NewReader(formatted), nil
	}
	return ret, nil
}
