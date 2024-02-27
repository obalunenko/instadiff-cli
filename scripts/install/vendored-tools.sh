#!/bin/bash

set -eu

SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(dirname "$0")"
REPO_ROOT="$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)"
SCRIPTS_DIR="${REPO_ROOT}/scripts"
source "${SCRIPTS_DIR}/helpers-source.sh"

echo "${SCRIPT_NAME} is running... "

export DOCKERFILE_PATH="${REPO_ROOT}/build/docker/go-tools/Dockerfile"

export IMAGE_NAME="${GOTOOLS_IMAGE:-${DOCKER_REPO}go-tools:${VERSION}}"

APP_NAME="go-tools"

echo "Building ${IMAGE_NAME} of ${APP_NAME} ..."

docker buildx bake -f "${REPO_ROOT}/build/docker/bake.hcl" "${APP_NAME}"