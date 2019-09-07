#!/usr/bin/env bash

# Get new tags from the remote
git fetch --tags

# Get the latest tag name
latestTag=$(git describe --tags $(git rev-list --tags --max-count=1))
echo ${latestTag}

curl -sL https://git.io/goreleaser | bash