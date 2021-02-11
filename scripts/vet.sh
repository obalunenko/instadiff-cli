#!/usr/bin/env sh

set -e

# shellcheck disable=SC2046
go vet $(go list ./...)

echo "Done."