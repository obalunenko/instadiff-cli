// +build tools

package tools

import (
	_ "github.com/axw/gocov/gocov"
	_ "github.com/go-bindata/go-bindata/v3/go-bindata"
	_ "github.com/golang/mock/mockgen"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/matm/gocov-html"
	_ "github.com/vasi-stripe/gogroup/cmd/gogroup"
	_ "golang.org/x/lint/golint"
	_ "golang.org/x/tools/cmd/cover"
	_ "golang.org/x/tools/cmd/stringer"
)
