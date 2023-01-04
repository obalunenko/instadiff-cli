#!/bin/bash

set -Eeuo pipefail

SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(dirname "$0")"
REPO_ROOT="$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)"
SCRIPTS_DIR="${REPO_ROOT}/scripts"
COVER_DIR=${REPO_ROOT}/coverage

source "${SCRIPTS_DIR}/helpers-source.sh"

echo "${SCRIPT_NAME} is running... "

checkInstalled 'coverbadger'
checkInstalled 'gocov'

COVERAGE=$(gocov report "${COVER_DIR}/full.json" | tail -1 | awk '{if ($1 != "?") print $3; else print "0.0";}' | sed 's/\%//g')
if [[ ${COVERAGE} == "NaN" ]]; then
  COVERAGE="0.0"
fi

coverbadger \
  --coverage="${COVERAGE}" \
  --md="${REPO_ROOT}/README.md" \
  --style=flat

echo "${SCRIPT_NAME} done."
