#!/usr/bin/env bash

if [[ -f "$(go env GOPATH)/bin/gogroup" ]] || [[ -f "/usr/local/bin/gogroup" ]]; then
  gogroup -order std,other,prefix=github.com/oleg-balunenko/ --rewrite $(find . -type f -name "*.go" | grep -v "vendor/" | grep -v ".git")
else
  printf "Cannot check gogroup, please run:
    go get -u -v github.com/Bubblyworld/gogroup/... \n"
  exit 1
fi
