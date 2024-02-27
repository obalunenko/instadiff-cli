#!/usr/bin/env bash

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly ROOT_DIR="$(dirname "$CURRENT_DIR")"
readonly GO_MOD_FILE="${ROOT_DIR}/go.mod"

function main() {
  echo "Updating Go version:"

  local currentGoVersion="$(extractCurrentVersion)"
  echo " - Current: ${currentGoVersion}"
  local escapedCurrentGoVersion="$(echo "${currentGoVersion}" | sed 's/\./\\./g')"

  local goVersion="${1}"
  local escapedGoVersion="$(echo "${goVersion}" | sed 's/\./\\./g')"
  echo " - New: ${goVersion}"

  # bump mod files in all the modules
  for modFile in $(find "${ROOT_DIR}" -name "go.mod" -not -path "${ROOT_DIR}/vendor/*" -not -path "${ROOT_DIR}/.git/*"); do
    bumpModFile "${modFile}" "${escapedCurrentGoVersion}" "${escapedGoVersion}"
  done

  # bump markdown files
  for f in $(find "${ROOT_DIR}" -name "*.md"); do
    bumpGolangDockerImages "${f}" "${escapedCurrentGoVersion}" "${escapedGoVersion}"
  done

  # bump github action workflows
  for f in $(find "${ROOT_DIR}/.github/workflows" -name "*.yml"); do
    bumpCIMatrix "${f}" "${escapedCurrentGoVersion}" "${escapedGoVersion}"
  done
}

# it will replace the 'go-version: [${oldGoVersion}, 1.x]' with 'go-version: [${newGoVersion}, 1.x]' in the given file
function bumpCIMatrix() {
  local file="${1}"
  local oldGoVersion="${2}"
  local newGoVersion="${3}"

    sed "s/go-version: \[${oldGoVersion}/go-version: \[${newGoVersion}/g" ${file} > ${file}.tmp
    mv ${file}.tmp ${file}
}

# it will replace the 'golang:${oldGoVersion}' with 'golang:${newGoVersion}' in the given file
function bumpGolangDockerImages() {
  local file="${1}"
  local oldGoVersion="${2}"
  local newGoVersion="${3}"

    sed "s/golang:${oldGoVersion}/golang:${newGoVersion}/g" ${file} > ${file}.tmp
    mv ${file}.tmp ${file}

}

# it will replace the 'go ${oldGoVersion}' with 'go ${newGoVersion}' in the given go.mod file
function bumpModFile() {
  local goModFile="${1}"
  local oldGoVersion="${2}"
  local newGoVersion="${3}"

    sed "s/^go ${oldGoVersion}/go ${newGoVersion}/g" ${goModFile} > ${goModFile}.tmp
    mv ${goModFile}.tmp ${goModFile}

}

# This function reads the reaper.go file and extracts the current version.
function extractCurrentVersion() {
  cat "${GO_MOD_FILE}" | grep '^go .*' | sed 's/^go //g' | head -n 1
}

main "$@"