#!/bin/sh

set -eu

SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(dirname "$0")"
REPO_ROOT="$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)"
BIN_DIR=${REPO_ROOT}/bin

echo "${SCRIPT_NAME} is running... "

APP=${APP_NAME}

echo "Building ${APP}..."

COMMIT="$(git rev-parse HEAD)"
SHORTCOMMIT="$(git rev-parse --short HEAD)"
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
VERSION="$(git tag --sort=committerdate | tail -1)"
GOVERSION="$(go version | awk '{print $3;}')"

if [ -z "${VERSION}" ] || [ "${VERSION}" = "${SHORTCOMMIT}" ]
 then
  VERSION="v0.0.0"
fi

BIN_OUT="${BIN_DIR}/${APP}"

BUILDINFO_VARS_PKG=github.com/obalunenko/version
GO_BUILD_LDFLAGS="-s -w \
-X ${BUILDINFO_VARS_PKG}.version=${VERSION} \
-X ${BUILDINFO_VARS_PKG}.commit=${COMMIT} \
-X ${BUILDINFO_VARS_PKG}.shortcommit=${SHORTCOMMIT} \
-X ${BUILDINFO_VARS_PKG}.builddate=${DATE} \
-X ${BUILDINFO_VARS_PKG}.appname=${APP} \
-X ${BUILDINFO_VARS_PKG}.goversion=${GOVERSION}"

GO_BUILD_PACKAGE="${REPO_ROOT}/cmd/${APP}"

rm -rf "${BIN_OUT}"

go build -o "${BIN_OUT}" -a -ldflags "${GO_BUILD_LDFLAGS}" "${GO_BUILD_PACKAGE}"

echo "Binary compiled at ${BIN_OUT}"
