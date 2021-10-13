#!/bin/sh

set -eu

SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(dirname "$0")"
REPO_ROOT="$(cd ${SCRIPT_DIR} && git rev-parse --show-toplevel)"
TOOLS_DIR=${REPO_ROOT}/tools

echo "${SCRIPT_NAME} is running... "

go env -w GOPROXY=https://goproxy.io,https://proxy.golang.org
go env -w GOPRIVATE=github.com/melsoft-games/bitech-go-shared
go env -w GONOSUMDB=github.com/melsoft-games/*

sync_vendor() {
  go mod tidy -v
  go mod vendor
  go mod verify
}

cd ${REPO_ROOT} || exit 1
pwd
sync_vendor

cd ${TOOLS_DIR} || exit 1
pwd
sync_vendor

echo "${SCRIPT_NAME} done."
