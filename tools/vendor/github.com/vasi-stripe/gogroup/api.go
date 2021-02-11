// Package gogroup helps to group import statements in Go files.
//
// Unlike goimports, it does not maintain existing import statement groups.
// Instead, it allows defining a canonical order of import statements, and fixes
// up files to enforce this order.
//
// Whatever order of import statements is used, the following rules are
// enforced:
//
// - Import statements within the same group have no empty lines between them.
// - Between two groups is an empty line.
// - Within a group, statements are sorted by path.
package gogroup

import "io"

// A Grouper determines groupings of import statements.
type Grouper interface {
	// Group determines the import group that an import statement should be in.
	//
	// The input is the package import path, eg: "os" or "github.com/example/repo".
	//
	// The output is the group number. If two import statements should be in the same
	// group, their group number should be identical. If an import statement should
	// be in a group after another import statement, its group number should be higher.
	// Otherwise, group numbers are arbitrary. Gaps between group numbers are explicitly
	// allowed.
	Group(pkgPath string) (group int)
}

// Processor processes files according to import grouping rules.
type Processor struct {
	grouper Grouper
}

// NewProcessor creates a new Processor with a given group definition.
func NewProcessor(grouper Grouper) *Processor {
	return &Processor{grouper}
}

// ValidationError is an error about incorrect import grouping.
type ValidationError struct {
	// Line is the line of the file at which the error occurred.
	Line int
	// ImportPath is the path being imported.
	ImportPath string
	// Message is a description of why this was an error.
	Message string
}

// Validate determines whether the existing import grouping of a source file is
// correct, according to our grouping rules.
//
// If the grouping is correct, Validate will return nil, nil.
// If an unexpected error occurs, Validate returns an error in err.
// Otherwise, if the grouping is incorrect, Validate returns an error in validErr.
//
// The fileName parameter is needed for error reporting only. You may leave it
// blank.
func (p *Processor) Validate(fileName string, r io.Reader) (validErr *ValidationError, err error) {
	return p.validate(fileName, r)
}

// Repair repairs the import grouping of a source file.
//
// If no repairs are necessary, a nil io.Reader will be returned. If repairs
// are needed, the new file content will be available in the returned
// io.Reader.
//
// The fileName parameter is needed for error reporting only. You may leave it
// blank.
func (p *Processor) Repair(fileName string, r io.Reader) (io.Reader, error) {
	return p.repair(fileName, r)
}

// Reformat both formats the file with goimports, and repairs any import groupings.
//
// The fileName is necessary for determining missing imports.
func (p *Processor) Reformat(fileName string, r io.Reader) (io.Reader, error) {
	return p.reformat(fileName, r)
}
