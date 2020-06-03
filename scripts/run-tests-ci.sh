#!/usr/bin/env bash
set -e

export GO111MODULE=on
go test -coverpkg=./... -covermode=atomic -coverprofile=coverage.out -json ./... > tests.out