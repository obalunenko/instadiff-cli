//go:build tools
// +build tools

package tools

import (
	_ "github.com/axw/gocov/gocov"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/goreleaser/goreleaser"
	_ "github.com/matm/gocov-html"
	_ "github.com/mattn/goveralls"
	_ "github.com/segmentio/golines"
	_ "github.com/vasi-stripe/gogroup/cmd/gogroup"
	_ "golang.org/x/lint/golint"
	_ "golang.org/x/tools/cmd/cover"
	_ "golang.org/x/tools/cmd/stringer"
	_ "gotest.tools/gotestsum"

	_ "github.com/obalunenko/coverbadger/cmd/coverbadger"
)
