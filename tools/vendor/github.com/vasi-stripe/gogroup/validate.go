package gogroup

import (
	"fmt"
	"io"
)

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s (line %s)", e.Message, e.ImportPath, e.Line)
}

// Yield a validation error.
func validationError(g *groupedImport, msg string) *ValidationError {
	return &ValidationError{
		Message:    msg,
		ImportPath: g.path,
		Line:       g.startLine,
	}
}

const (
	errstrStatementOrder     = "Import out of order within import group"
	errstrStatementExtraLine = "Extra empty line inside import group"
	errstrStatementGroup     = "Import in incorrect group"
	errstrGroupOrder         = "Import groups out of order"
	errstrGroupExtraLine     = "Extra empty line between import groups"
)

// Validate an import group.
func (gs groupedImports) validate() *ValidationError {
	if len(gs) < 2 {
		// Always valid!
		return nil
	}

	var prev *groupedImport
	for _, g := range gs {
		if prev != nil {
			emptyLines := g.startLine - prev.endLine - 1

			if g.group == prev.group {
				if emptyLines > 0 {
					return validationError(g, errstrStatementExtraLine)
				} else if g.path < prev.path {
					return validationError(g, errstrStatementOrder)
				}
			} else if emptyLines == 0 {
				// This could also be a missing empty line.
				return validationError(g, errstrStatementGroup)
			} else if g.group < prev.group {
				return validationError(g, errstrGroupOrder)
			} else if emptyLines > 1 {
				return validationError(g, errstrGroupExtraLine)
			}

		}
		prev = g
	}
	return nil
}

// Validate a file.
func (p *Processor) validate(fileName string, r io.Reader) (validErr *ValidationError, err error) {
	gs, err := p.readImports(fileName, r)
	if err != nil {
		return nil, err
	}
	return gs.validate(), nil
}
