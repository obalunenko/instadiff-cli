#!/usr/bin/env bash
set -e

gofmt -s -w -l $(find . -type f -name '*.go' | grep -v 'vendor' |grep -v '.git' )

echo "Done."
