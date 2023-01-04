#!/bin/bash

set -e

SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(dirname "$0")"
REPO_ROOT="$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)"
SCRIPTS_DIR="${REPO_ROOT}/scripts"

source "${SCRIPTS_DIR}/helpers-source.sh"

echo "${SCRIPT_NAME} is running... "

checkInstalled 'goimports'

echo "Making filelist"
FILES=( $(find . -type f -name "*.go" -not -path "./vendor/*" -not -path "./tools/vendor/*" -not -path "./.git/*") )

LOCAL_PFX=$(go list -m)
echo "Local packages prefix: ${LOCAL_PFX}"

for f in "${FILES[@]}"; do
  echo "Fixing imports at ${f}"
  sed -i -- '/^import (/,/)/ {;/^$/ d;}' "$f"
  goimports -local=${LOCAL_PFX} -w "$f"
done


TORM=( $(find . -type f -name "*.go--" -not -path "./vendor/*" -not -path "./tools/vendor/*" -not -path "./.git/*") )

for f in "${TORM[@]}"; do
  rm -rf ${f}
done

echo "${SCRIPT_NAME} done."
