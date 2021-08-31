#!/bin/bash

set -Eeuo pipefail

SCRIPT_NAME="$(basename "$0")"

echo "${SCRIPT_NAME} is running... "

if [[ ! -f "$(go env GOPATH)/bin/golangci-lint" ]] && [[ ! -f "/usr/local/bin/golangci-lint" ]]; then
  echo "Install golangci-lint"
  echo "run 'make install-tools' "
  exit 1
fi

echo "Linting..."

golangci-lint run --no-config --disable-all -E govet
golangci-lint run --new-from-rev=HEAD~ --config .golangci.pipe.yml

echo "${SCRIPT_NAME} done."