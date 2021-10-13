#!/usr/bin/env bash
set -e

gofmt -s -w $(find . -type f -name '*.go' | grep -v 'vendor' |grep -v '.git' )

echo "Done."
