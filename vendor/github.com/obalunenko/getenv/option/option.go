// Package option provides options for parsing environment variables.
package option

import (
	"github.com/obalunenko/getenv/internal"
)

// Option implements options for EnvOrDefault.
type Option interface {
	Apply(params *internal.Parameters)
}

type withSeparator string

func (w withSeparator) Apply(p *internal.Parameters) {
	p.Separator = string(w)
}

// WithSeparator ads slice separator option.
func WithSeparator(separator string) Option {
	return withSeparator(separator)
}

type withTimeLayout string

func (w withTimeLayout) Apply(p *internal.Parameters) {
	p.Layout = string(w)
}

// WithTimeLayout ads time layout option.
func WithTimeLayout(layout string) Option {
	return withTimeLayout(layout)
}
