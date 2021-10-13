#!/bin/bash

set -Eeuo pipefail

function checkInstalled() {
  CMD=$1

  if ! command -v "${CMD}" &>/dev/null; then
    echo "Install ${CMD}"
    echo "run 'make install-tools' "
    exit 1
  fi
}

function openSource() {
  OPEN_CMD=""
  OPEN_SRC=$1

  case "$OSTYPE" in
  darwin*)
    OPEN_CMD=open
    ;;
  linux*)
    OPEN_CMD=xdg-open
    ;;
  msys* | cygwin*)
    OPEN_CMD=start
    ;;
  *)
    echo "unknown: $OSTYPE"
    exit 1
    ;;
  esac

  ${OPEN_CMD} "${OPEN_SRC}"
}
