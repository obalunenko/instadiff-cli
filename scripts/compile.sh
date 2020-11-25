#!/usr/bin/env sh
set -e
echo "Building..."

REPO_ROOT=$(git rev-parse --show-toplevel)
APP="instadiff-cli"
MODULE="github.com/obalunenko/instadiff-cli"
VERSION=$(git describe --tags "$(git rev-list --tags --max-count=1)")"-local"
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null)
DATE=$(date "+%Y-%m-%d")

GO_BUILD_LDFLAGS="-s -w -X 'main.commit=${COMMIT_HASH}' -X 'main.date=${DATE}' -X 'main.version=${VERSION}'"
GO_BUILD_PACKAGE="${MODULE}/cmd/instadiff-cli/."

BIN_OUT=${REPO_ROOT}/bin/${APP}

go build -o "${BIN_OUT}" -a -ldflags "${GO_BUILD_LDFLAGS}" ${GO_BUILD_PACKAGE}

echo "Binary compiled at ${BIN_OUT}"