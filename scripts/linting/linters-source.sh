#!/bin/bash

set -Eeuo pipefail

SCRIPT_DIR="$(dirname "$0")"
REPO_ROOT="$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)"
SCRIPTS_DIR="${REPO_ROOT}/scripts"

source "${SCRIPTS_DIR}/helpers-source.sh"

function vet() {
  echo "vet project..."
  declare -a vet_errs=$(go vet $(go list ./...) 2>&1 >/dev/null)
  EXIT_CODE=$?
  if [[ ${EXIT_CODE} -ne 0 ]]; then
    exit 1
  fi
  if [[ ${vet_errs} ]]; then
    echo "fix it:"
    for f in "${vet_errs[@]}"; do
      echo "$f"

    done
    exit 1

  else
    echo "code is ok"
    echo ${vet_errs}
  fi
  echo ""
  echo ""
}

function fmt() {
  echo "fmt lint..."

  checkInstalled 'gofmt'

  declare -a fmts=$(gofmt -s -l $(find . -type f -name '*.go' | grep -v 'vendor' | grep -v '.git'))

  if [[ ${fmts} ]]; then
    echo "fix it:"
    for f in "${fmts[@]}"; do
      echo "$f"

    done
    exit 1

  else
    echo "code is ok"
    echo ${fmts}
  fi
  echo ""
}

function go-lint() {
  echo "golint..."

  checkInstalled 'golint'

  declare -a lints=$(golint $(go list ./...)) ## its a hack to not lint generated code
  if [[ ${lints} ]]; then
    echo "fix it:"
    for l in "${lints[@]}"; do
      echo "$l"

    done
    exit 1

  else
    echo "code is ok"
    echo ${lints}
  fi

  echo ""
}

function go-group() {
  echo "goimports..."

  checkInstalled 'goimports'

  declare -a lints=$(goimports -l -local=$(go list -m) $(find . -type f -name "*.go" | grep -v "vendor/"))

  if [[ ${lints} ]]; then
    echo "fix it:"
    for l in "${lints[@]}"; do
      echo "$l"

    done
    exit 1

  else
    echo "code is ok"
    echo ${lints}
  fi
  echo ""

}

function golangci() {
  echo "golangci-lint linter running..."

  checkInstalled 'golangci-lint'

  golangci-lint run --config .golangci.yml ./...

  echo ""
}

function golangci-ci_execute() {
  echo "golangci-lint-ci_execute linter running..."

  checkInstalled 'golangci-lint'

  golangci-lint run ./... >linters.out

  echo ""
}
