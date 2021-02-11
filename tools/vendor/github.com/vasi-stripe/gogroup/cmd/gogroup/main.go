package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/vasi-stripe/gogroup"
)

type grouper struct {
	// The group numbers of prefixed packages.
	prefixes map[int]string

	// The group numbers of standard packages and unidentified packages.
	std, other int

	// The next integer to assign
	next int
}

func newGrouper() *grouper {
	return &grouper{
		prefixes: make(map[int]string),
		std:      0,
		other:    1,
		next:     2,
	}
}

func (g *grouper) Group(pkg string) int {
	for n, prefix := range g.prefixes {
		if strings.HasPrefix(pkg, prefix) {
			return n
		}
	}

	// A dot distinguishes non-standard packages.
	if strings.Contains(pkg, ".") {
		return g.other
	}

	return g.std
}

func (g *grouper) wasSet() bool {
	return g.next > 2
}

func (g *grouper) String() string {
	parts := []string{}
	remain := len(g.prefixes)
	for i := 0; i <= g.std || i <= g.other || remain > 0; i++ {
		if g.std == i {
			parts = append(parts, "std")
		} else if g.other == i {
			parts = append(parts, "other")
		} else if p, ok := g.prefixes[i]; ok {
			parts = append(parts, fmt.Sprintf("prefix=%s", p))
			remain--
		}
	}
	return strings.Join(parts, ",")
}

var rePrefix = regexp.MustCompile(`^prefix=(.*)$`)

func (g *grouper) Set(s string) error {
	parts := strings.Split(s, ",")
	for _, p := range parts {
		if p == "std" {
			g.std = g.next
		} else if p == "other" {
			g.other = g.next
		} else if match := rePrefix.FindStringSubmatch(p); match != nil {
			g.prefixes[g.next] = match[1]
		} else {
			return fmt.Errorf("Unknown order specification '%s'", p)
		}
		g.next++
	}
	return nil
}

const (
	statusError       = 1
	statusHelp        = 2
	statusInvalidFile = 3
)

func validateOne(proc *gogroup.Processor, file string) (validErr *gogroup.ValidationError, err error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return proc.Validate(file, f)
}

func validateAll(proc *gogroup.Processor, files []string) {
	invalid := false
	for _, file := range files {
		validErr, err := validateOne(proc, file)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(statusError)
		}
		if validErr != nil {
			invalid = true
			fmt.Fprintf(os.Stdout, "%s:%d: %s at %s\n", file, validErr.Line,
				validErr.Message, strconv.Quote(validErr.ImportPath))
		}
	}

	if invalid {
		os.Exit(statusInvalidFile)
	}
}

func rewriteOne(proc *gogroup.Processor, file string) error {
	// Get the rewritten file.
	r, err := func() (io.Reader, error) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return proc.Reformat(file, f)
	}()
	if err != nil {
		return err
	}

	if r != nil {
		// Write the result.
		f, err := os.Create(file)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, r)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Fixed %s\n", file)
	}
	return nil
}

func rewriteAll(proc *gogroup.Processor, files []string) {
	for _, file := range files {
		err := rewriteOne(proc, file)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(statusError)
		}
	}
}

func main() {
	rewrite := false
	gr := newGrouper()

	flag.Usage = func() {
		// Hard to get flag to format long usage well, so just put everything here.
		fmt.Fprintln(os.Stderr,
			`group-imports: Enforce import grouping in Go source files.

Exits with status 3 if import grouping is violated.

Usage: group-imports [OPTIONS] FILE...

  -rewrite
      Instead of checking import grouping, rewrite the source files with
      the correct grouping. Default: false.

  -order SPEC[,SPEC...]
      Modify the import grouping strategy by listing the desired groups in
      order. Group specifications include:

      - std: Standard library imports
      - prefix=PREFIX: Imports whose path starts with PREFIX
      - other: Imports that match no other specification

      These groups can be specified in one comma-separated argument, or
      multiple arguments. Default: std,other
`,
		)
	}

	flag.BoolVar(&rewrite, "rewrite", false, "")
	flag.Var(gr, "order", "")

	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "No file provided.")
		flag.Usage()
		os.Exit(statusHelp)
	}

	proc := gogroup.NewProcessor(gr)
	if rewrite {
		rewriteAll(proc, flag.Args())
	} else {
		validateAll(proc, flag.Args())
	}
}
