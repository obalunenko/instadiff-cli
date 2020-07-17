#!/usr/bin/env bash
set -e

function get_dependencies() {
  cd .tools || exit

  declare -a packages=(
    "golang.org/x/tools/cmd/cover/..."
    "github.com/mattn/goveralls/..."
    "github.com/Bubblyworld/gogroup/..."
    "golang.org/x/lint/golint"
    "golang.org/x/tools/cmd/stringer"
    "mvdan.cc/gofumpt/..."
  )

  ## now loop through the above array
  for pkg in "${packages[@]}"; do
    echo "$pkg"
    go get -u -v "$pkg"
  done

  curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin

  cd - || exit
}

echo Gonna to update go tools and packages...
get_dependencies
echo All is done!
