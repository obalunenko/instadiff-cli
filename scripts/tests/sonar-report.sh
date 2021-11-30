#!/bin/sh

set -eu pipefail

SCRIPT_NAME="$(basename "$0")"

echo "${SCRIPT_NAME} is running... "

go test -json ./... > tests-report.json

echo "${SCRIPT_NAME} done."
