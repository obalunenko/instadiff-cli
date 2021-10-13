#!/bin/bash

set -Eeuo pipefail

SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(dirname "$0")"
REPO_ROOT="$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)"
SCRIPTS_DIR="${REPO_ROOT}/scripts"
COVER_DIR=${REPO_ROOT}/coverage

source "${SCRIPTS_DIR}/helpers-source.sh"

checkInstalled gocov

echo "${SCRIPT_NAME} is running... "

export GO111MODULE=on

COVER_DIR=.cover
COVER_FULL=.cover/full.cov
COVER_REPORT=cover.out


rm -rf ${COVER_DIR} ${COVER_REPORT}
mkdir -p ${COVER_DIR}

go test --count=1 -coverprofile ${COVER_DIR}/unit.cov -covermode=atomic ./...


{
echo "mode: atomic"
tail -q -n +2 ${COVER_DIR}/*.cov
} >> ${COVER_FULL}

gocov convert "${COVER_FULL}" >"${COVER_DIR}/full.json"

mv ${COVER_FULL} ${COVER_REPORT}

echo "${SCRIPT_NAME} done."