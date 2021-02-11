#!/usr/bin/env sh

set -e

go vet $(go list ./...)

echo "Done."