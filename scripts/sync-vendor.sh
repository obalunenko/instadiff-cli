#!/bin/sh

set -eu

SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(dirname "$0")"
REPO_ROOT="$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)"

echo "${SCRIPT_NAME} is running... "

sync_vendor() {
  go mod tidy -v
  go mod vendor
  go mod verify
}

cd "${REPO_ROOT}" || exit 1
pwd
sync_vendor

echo "${SCRIPT_NAME} done."
