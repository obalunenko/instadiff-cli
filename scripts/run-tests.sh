#!/usr/bin/env bash
set -e

export GO111MODULE=on
go test -v -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.out ./...
