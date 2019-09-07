#!/usr/bin/env bash

REPO_ROOT=$(git rev-parse --show-toplevel)
SCRIPTS_DIR=${REPO_ROOT}/scripts

source ${SCRIPTS_DIR}/linters.sh

vet
fmt
go-lint
go-group
golangci
