#!/usr/bin/env bash

REPO_ROOT=$(git rev-parse --show-toplevel)
SCRIPTS_DIR=${REPO_ROOT}/scripts

# shellcheck disable=SC1090
source "${SCRIPTS_DIR}"/linters.sh

vet
fmt
go-lint
go-group
golangci-ci_execute
