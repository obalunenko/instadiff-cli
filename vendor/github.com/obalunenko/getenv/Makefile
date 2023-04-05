.PHONY: help, test-cover, prepare-cover-report, open-cover-report, update-readme-cover, update-readme-doc, test, test-regression, configure, sync-vendor, imports, fmt, format-project, install-tools, vet, lint-full, lint-pipeline, lint-sonar, release, release-local-snapshot, check-releaser, new-version
BIN_DIR=./bin

SHELL := env VERSION=$(VERSION) $(SHELL)
VERSION ?= $(shell git describe --tags $(git rev-list --tags --max-count=1))

APP_NAME?=getenv
SHELL := env APP_NAME=$(APP_NAME) $(SHELL)

GOTOOLS_IMAGE_TAG?=v0.6.1
SHELL := env GOTOOLS_IMAGE_TAG=$(GOTOOLS_IMAGE_TAG) $(SHELL)

COMPOSE_TOOLS_FILE=deployments/docker-compose/go-tools-docker-compose.yml
COMPOSE_TOOLS_CMD_BASE=docker compose -f $(COMPOSE_TOOLS_FILE)
COMPOSE_TOOLS_CMD_UP=$(COMPOSE_TOOLS_CMD_BASE) up --exit-code-from
COMPOSE_TOOLS_CMD_PULL=$(COMPOSE_TOOLS_CMD_BASE) pull

TARGET_MAX_CHAR_NUM=20

## Show help
help:
	${call colored, help is running...}
	@echo ''
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  %-$(TARGET_MAX_CHAR_NUM)s %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)


## Test coverage report.
test-cover:
	./scripts/tests/coverage.sh

prepare-cover-report: test-cover
	$(COMPOSE_TOOLS_CMD_UP) prepare-cover-report prepare-cover-report

## Open coverage report.
open-cover-report: prepare-cover-report
	./scripts/open-coverage-report.sh

## Update readme coverage.
update-readme-cover: prepare-cover-report
	$(COMPOSE_TOOLS_CMD_UP) update-readme-coverage update-readme-coverage

## Update readme doc.
update-readme-doc:
	$(COMPOSE_TOOLS_CMD_UP) update-readme-doc update-readme-doc

## Run tests.
test:
	$(COMPOSE_TOOLS_CMD_UP) run-tests run-tests

## Run regression tests.
test-regression: test

## Sync vendor and install needed tools.
configure: sync-vendor install-tools

## Sync vendor with go.mod.
sync-vendor:
	./scripts/sync-vendor.sh

## Fix imports sorting.
imports:
	$(COMPOSE_TOOLS_CMD_UP) fix-imports fix-imports

## Format code with go fmt.
fmt:
	$(COMPOSE_TOOLS_CMD_UP) fix-fmt fix-fmt

## Format code and sort imports.
format-project: fmt imports

## Installs vendored tools.
install-tools:
	echo "Installing ${GOTOOLS_IMAGE_TAG}"
	$(COMPOSE_TOOLS_CMD_PULL)

## vet project
vet:
	./scripts/linting/run-vet.sh

## Run full linting
lint-full:
	$(COMPOSE_TOOLS_CMD_UP) lint-full lint-full

## Run linting for build pipeline
lint-pipeline:
	$(COMPOSE_TOOLS_CMD_UP) lint-pipeline lint-pipeline

## Run linting for sonar report
lint-sonar:
	$(COMPOSE_TOOLS_CMD_UP) lint-sonar lint-sonar

## Release
release:
	./scripts/release/release.sh

## Release local snapshot
release-local-snapshot:
	./scripts/release/local-snapshot-release.sh

## Check goreleaser config.
check-releaser:
	./scripts/release/check.sh

## Issue new release.
new-version: vet test-regression
	./scripts/release/new-version.sh


.DEFAULT_GOAL := help

