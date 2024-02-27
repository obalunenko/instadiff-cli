#!/bin/bash

set -eu

SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(dirname "$0")"
REPO_ROOT="$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)"
SCRIPTS_DIR="${REPO_ROOT}/scripts"
source "${SCRIPTS_DIR}/helpers-source.sh"

BIN_DIR=${REPO_ROOT}/bin
mkdir -p "${BIN_DIR}"

echo "${SCRIPT_NAME} is running... "

APP=${APP_NAME}

echo "Building ${APP}..."

COMMIT="$(git rev-parse HEAD)"
SHORTCOMMIT="$(git rev-parse --short HEAD)"
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
VERSION="$(git tag | sort -V | tail -1)"
GOVERSION="$(go version | awk '{print $3;}')"

if [ -z "${VERSION}" ] || [ "${VERSION}" = "${SHORTCOMMIT}" ]; then
  VERSION="v0.0.0"
fi

COMMIT_TAG=""

if [[ "${VERSION}" != "v0.0.0" ]]; then
  COMMIT_TAG="$(git rev-list -n 1 ${VERSION})"
fi

## check if the version is a tag
if [[ "${COMMIT_TAG}" != "${COMMIT}" ]]; then
  echo 'dev'

  VERSION="${VERSION}-dev"
fi

## check if there are uncommitted changes
if [[ $(git diff --stat) != '' ]]; then
  echo 'dirty'

  COMMIT="${COMMIT}-dirty"
  SHORTCOMMIT="${SHORTCOMMIT}-dirty"
  VERSION="${VERSION}-dirty"
fi

BIN_OUT="${BIN_DIR}/${APP}"

BUILDINFO_VARS_PKG=github.com/obalunenko/version
export GO_BUILD_LDFLAGS="-s -w \
-X ${BUILDINFO_VARS_PKG}.version=${VERSION} \
-X ${BUILDINFO_VARS_PKG}.commit=${COMMIT} \
-X ${BUILDINFO_VARS_PKG}.shortcommit=${SHORTCOMMIT} \
-X ${BUILDINFO_VARS_PKG}.builddate=${DATE} \
-X ${BUILDINFO_VARS_PKG}.appname=${APP} \
-X ${BUILDINFO_VARS_PKG}.goversion=${GOVERSION}"

GO_BUILD_PACKAGE="${REPO_ROOT}/cmd/${APP}"

rm -rf "${BIN_OUT}"

go build -trimpath -o "${BIN_OUT}" -a -ldflags "${GO_BUILD_LDFLAGS}" "${GO_BUILD_PACKAGE}"

echo "Build ${BIN_OUT} success"
