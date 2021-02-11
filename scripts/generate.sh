#!/usr/bin/env sh
set -e
echo "Recreate generated code ...."

SCRIPT_NAME="$(basename "$(test -L "$0" && readlink "$0" || echo "$0")")"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd ${SCRIPT_DIR} && git rev-parse --show-toplevel)"
BIN_DIR=${ROOT_DIR}/bin

cd ${ROOT_DIR} || exit 1

go generate ./...

cd - || exit 1

echo "Done."