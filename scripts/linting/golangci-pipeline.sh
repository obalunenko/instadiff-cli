#!/bin/bash

set -Eeuo pipefail

SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(dirname "$0")"
REPO_ROOT="$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)"
SCRIPTS_DIR="${REPO_ROOT}/scripts"

echo "${SCRIPT_NAME} is running... "

source "${SCRIPTS_DIR}/linting/linters-source.sh"

checkInstalled golangci-lint

echo "Linting..."

golangci-lint run --out-format=github-actions --no-config --disable-all -E govet
golangci-lint run --config .golangci.pipe.yml

echo "${SCRIPT_NAME} done."
