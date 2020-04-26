#!/usr/bin/env bash
set -e

REPO_ROOT=$(git rev-parse --show-toplevel)
APP="instadiff-cli"
MODULE="github.com/oleg-balunenko/instadiff-cli"
VERSION=$(git describe --tags "$(git rev-list --tags --max-count=1)")"-local"
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null)
DATE=$(date "+%Y-%m-%d")
BUILD_PLATFORM=$(uname -a | awk '{print tolower($1);}')
IMPORT_DURING_SOLVE=${IMPORT_DURING_SOLVE:-false}

if [[ "$(pwd)" != "${REPO_ROOT}" ]]; then
  echo "you are not in the root of the repo" 1>&2
  echo "please cd to ${REPO_ROOT} before running this script" 1>&2
  exit 1
fi

GO_BUILD_LDFLAGS="-s -w -X 'main.commit=${COMMIT_HASH}' -X 'main.date=${DATE}' -X 'main.version=${VERSION}'"
GO_BUILD_PACKAGE="${MODULE}/cmd/instadiff-cli/."

if [[ -z "${BUILD_PLATFORMS}" ]]; then
  BUILD_PLATFORMS="linux windows darwin"
fi

if [[ -z "${BUILD_ARCHS}" ]]; then
  BUILD_ARCHS="amd64 386"
fi

mkdir -p "${REPO_ROOT}/release"

for OS in ${BUILD_PLATFORMS[@]}; do
  for ARCH in ${BUILD_ARCHS[@]}; do
    NAME="${APP}-${OS}-${ARCH}"
    if [[ "${OS}" == "windows" ]]; then
      NAME="${NAME}.exe"
    fi

    if [[ "${OS}" == "darwin" && "${BUILD_PLATFORM}" == "darwin" ]]; then
      CGO_ENABLED=0
    else
      CGO_ENABLED=0
    fi
    if [[ "${ARCH}" == "ppc64" || "${ARCH}" == "ppc64le" ]] && [[ "${OS}" != "linux" ]]; then
      # ppc64 and ppc64le are only supported on Linux.
      echo "Building for ${OS}/${ARCH} not supported."
    else
      echo "Building for ${OS}/${ARCH} with CGO_ENABLED=${CGO_ENABLED}"
      GOARCH=${ARCH} GOOS=${OS} CGO_ENABLED=${CGO_ENABLED}
      go build -o "${REPO_ROOT}"/release/${NAME} -a -ldflags "${GO_BUILD_LDFLAGS}" ${GO_BUILD_PACKAGE}

      pushd "${REPO_ROOT}/release" >/dev/null
      shasum -a 256 "${NAME}" >"${NAME}.sha256"
      popd >/dev/null
    fi
  done
done
