#!/usr/bin/env bash

export GO111MODULE=on
go test -v -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.out ./...
