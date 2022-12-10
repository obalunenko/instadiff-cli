#!/bin/bash

set -e

SCRIPT_NAME="$(basename "$0")"

echo "${SCRIPT_NAME} is running... "

GOTEST="go test -v "
if command -v "gotestsum" &>/dev/null; then
  GOTEST="gotestsum --format pkgname-and-test-fails --"
fi

${GOTEST} -race ./...

echo "${SCRIPT_NAME} done."
