#!/usr/bin/env bash

set -Eeuo pipefail

function cleanup() {
  trap - SIGINT SIGTERM ERR EXIT
  echo "cleanup running"
}

trap cleanup SIGINT SIGTERM ERR EXIT

SCRIPT_NAME="$(basename "$(test -L "$0" && readlink "$0" || echo "$0")")"

echo "${SCRIPT_NAME} is running... "

function require_clean_work_tree() {
  # Update the index
  git update-index -q --ignore-submodules --refresh
  err=0

  # Disallow unstagged changes in the working tree
  if ! git diff-files --quiet --ignore-submodules --; then
    echo >&2 "cannot $1: you have unstaged changes."
    git diff-files --name-status -r --ignore-submodules -- >&2
    err=1
  fi

  # Disallow uncommitted changes in the index
  if ! git diff-index --cached --quiet HEAD --ignore-submodules --; then
    echo >&2 "cannot $1: your index contains uncommitted changes."
    git diff-index --cached --name-status -r --ignore-submodules HEAD -- >&2
    err=1
  fi

  if [[ ${err} == 1 ]]; then
    echo >&2 "Please commit or stash them."
    exit 1
  fi
}

function menu() {
  clear
  printf "Select what you want to update: \n"
  printf "1 - Major update\n"
  printf "2 - Minor update\n"
  printf "3 - Patch update\n"
  printf "4 - Exit\n"
  read -r selection

  case "$selection" in
  1)
    printf "Major updates......\n"
    NEW_VERSION=$(git tag | sed 's/\(.*v\)\([0-9]*\)\.\([0-9]*\)\.\([0-9]*\)/\2;\3;\4;\1/g' | sort -t';' -k 1,1n -k 2,2n -k 3,3n | tail -n 1 | awk -F';' '{printf "%s%d.%d.%d", $4, ($1+1),0,0 }')
    ;;
  2)
    printf "Run Minor update.........\n"
    NEW_VERSION=$(git tag | sed 's/\(.*v\)\([0-9]*\)\.\([0-9]*\)\.\([0-9]*\)/\2;\3;\4;\1/g' | sort -t';' -k 1,1n -k 2,2n -k 3,3n | tail -n 1 | awk -F';' '{printf "%s%d.%d.%d", $4, $1,($2+1),0 }')
    ;;
  3)
    printf "Patch update.........\n"
    NEW_VERSION=$(git tag | sed 's/\(.*v\)\([0-9]*\)\.\([0-9]*\)\.\([0-9]*\)/\2;\3;\4;\1/g' | sort -t';' -k 1,1n -k 2,2n -k 3,3n | tail -n 1 | awk -F';' '{printf "%s%d.%d.%d", $4, $1,$2,($3 + 1) }')
    ;;
  4)
    printf "Exit................................\n"
    exit 1
    ;;
  *)
    clear
    printf "Incorrect selection. Try again\n"
    menu
    ;;
  esac

}

## Check if git is clean
require_clean_work_tree "create new version"

git pull

## Sem ver update menu
menu

NEW_TAG=${NEW_VERSION}

TAG_COMMIT=$(git rev-list --tags --max-count=1)
CURRENT_TAG=$(git describe --tags "${TAG_COMMIT}")
CHANGELOG="$(git log --pretty=format:"%s" HEAD..."${CURRENT_TAG}")"


echo "New version is: ${NEW_TAG}"
while true; do
  echo "Is it ok? (:y)?:"
  read -r yn
  case $yn in
  [Yy]*)

    git tag -a "${NEW_TAG}" -m "${CHANGELOG}" && \
     git push --tags

    break
    ;;
  [Nn]*)
    echo "Cancel"
    break
    ;;
  *)
    echo "Please answer yes or no."
    ;;
  esac
done

echo "${SCRIPT_NAME} done."
