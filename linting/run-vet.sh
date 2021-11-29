#!/bin/bash

set -Eeuo pipefail

SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(dirname "$0")"

echo "${SCRIPT_NAME} is running... "

source "${SCRIPT_DIR}/linters-source.sh"

vet

echo "${SCRIPT_NAME} done."
