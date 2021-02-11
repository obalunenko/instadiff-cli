#!/usr/bin/env bash

function check_status() {
  # first param is error message to print in case of error
  if [ "$?" -ne 0 ]
    then
      if [ -n "$1" ]
	    then
          echo "$1"
      fi
      exit 1
  fi
}

function run_go_install_in_parallel() {
  cd ./tools || exit 1
	# params are installing go apps
	apps=()
	export GO111MODULE="on"
	for app in "$@"
	do
	    if [ -d "${app}" ]
		then
		  apps+=("${app}/...")
		  echo "[INFO]: Going to build $app binary..."
		else
		  echo "[WARN]: $app not found, skipping..."
	     fi
	done
    go install -mod=vendor "${apps[@]}"
    export GO111MODULE="auto"
    check_status "[FAIL]: build failed!"
    echo "[SUCCESS]: build finished."
}


run_go_install_in_parallel \
"./vendor/github.com/golangci/golangci-lint/cmd/golangci-lint" \
"./vendor/github.com/vasi-stripe/gogroup/cmd/gogroup" \
"./vendor/github.com/matm/gocov-html" \
"./vendor/github.com/axw/gocov/gocov" \
"./vendor/golang.org/x/tools/cmd/stringer" \
"./vendor/github.com/golang/mock/mockgen" \
"./vendor/github.com/go-bindata/go-bindata/v3/go-bindata" \
"./vendor/golang.org/x/lint/golint" \
"./vendor/github.com/mattn/goveralls"