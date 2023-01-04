#!/bin/bash

set -Eeuo pipefail

SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(dirname "$0")"
REPO_ROOT="$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)"
SCRIPTS_DIR="${REPO_ROOT}/scripts"

source "${SCRIPTS_DIR}/helpers-source.sh"

echo "${SCRIPT_NAME} is running... "

checkInstalled 'goreleaser'

APP=${APP_NAME}

echo "${SCRIPT_NAME} is running fo ${APP}... "

# Get new tags from the remote
git fetch --tags -f

COMMIT="$(git rev-parse HEAD)"
SHORTCOMMIT="$(git rev-parse --short HEAD)"
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
VERSION="$(git tag --sort=committerdate | tail -1)"
GOVERSION="$(go version | awk '{print $3;}')"

if [ -z "${VERSION}" ] || [ "${VERSION}" = "${SHORTCOMMIT}" ]
 then
  VERSION="v0.0.0"
fi

BUILDINFO_VARS_PKG=github.com/obalunenko/version
export GO_BUILD_LDFLAGS="-s -w \
-X ${BUILDINFO_VARS_PKG}.version=${VERSION} \
-X ${BUILDINFO_VARS_PKG}.commit=${COMMIT} \
-X ${BUILDINFO_VARS_PKG}.shortcommit=${SHORTCOMMIT} \
-X ${BUILDINFO_VARS_PKG}.builddate=${DATE} \
-X ${BUILDINFO_VARS_PKG}.appname=${APP} \
-X ${BUILDINFO_VARS_PKG}.goversion=${GOVERSION}"

goreleaser release --rm-dist
