#!/bin/bash

set -Eeuo pipefail

SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(dirname "$0")"
REPO_ROOT="$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)"
SCRIPTS_DIR="${REPO_ROOT}/scripts"

echo "${SCRIPT_NAME} is running... "

source "${SCRIPTS_DIR}/linting/linters-source.sh"

vet
fmt
go-group
golangci

echo "${SCRIPT_NAME} done."
