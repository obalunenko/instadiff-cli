// Package option provides options for parsing environment variables.
package option

import (
	"github.com/obalunenko/getenv/internal"
)

// Option is a contract for EnvParser parameters options.
type Option interface {
	// Apply applies the option to the parameters.
	Apply(params *internal.Parameters)
}

type withSeparator string

func (w withSeparator) Apply(p *internal.Parameters) {
	p.Separator = string(w)
}

// WithSeparator adds slice separator option.
func WithSeparator(separator string) Option {
	return withSeparator(separator)
}

type withTimeLayout string

func (w withTimeLayout) Apply(p *internal.Parameters) {
	p.Layout = string(w)
}

// WithTimeLayout adds time.Time layout option.
func WithTimeLayout(layout string) Option {
	return withTimeLayout(layout)
}
