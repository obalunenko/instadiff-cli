#!/bin/bash

set -Eeuo pipefail

SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(dirname "$0")"
REPO_ROOT="$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)"
SCRIPTS_DIR="${REPO_ROOT}/scripts"

source "${SCRIPTS_DIR}/helpers-source.sh"

echo "${SCRIPT_NAME} is running... "

checkInstalled 'gofmt'

echo "Making filelist"
GO_FILES=( $(find . -type f -name "*.go" -not -path "./vendor/*" -not -path "./tools/vendor/*" -not -path "./.git/*") )

LOCAL_PFX=$(go list -m)
echo "Local packages prefix: ${LOCAL_PFX}"

for f in "${GO_FILES[@]}"; do
  echo "Fixing fmt at ${f}"
  gofmt -s -w "$f"
done

echo "${SCRIPT_NAME} done."
