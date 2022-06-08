#!/bin/bash

set -Eeuo pipefail

SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(dirname "$0")"
REPO_ROOT="$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)"
SCRIPTS_DIR="${REPO_ROOT}/scripts"

source "${SCRIPTS_DIR}/helpers-source.sh"

echo "${SCRIPT_NAME} is running... "

checkInstalled 'gofmt'

GO_FILES=$(find . -type f -name '*.go' | grep -v 'vendor' | grep -v '.git')

gofmt -s -w ${GO_FILES}

echo "${SCRIPT_NAME} done."
