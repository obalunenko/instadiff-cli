BIN_DIR=./bin

SHELL := env VERSION=$(VERSION) $(SHELL)
VERSION ?= $(shell git describe --tags $(git rev-list --tags --max-count=1))

APP_NAME?=instadiff-cli
SHELL := env APP_NAME=$(APP_NAME) $(SHELL)


GOTOOLS_IMAGE_TAG?=v0.4.3
SHELL := env GOTOOLS_IMAGE_TAG=$(GOTOOLS_IMAGE_TAG) $(SHELL)

COMPOSE_TOOLS_FILE=deployments/docker-compose/go-tools-docker-compose.yml
COMPOSE_TOOLS_CMD_BASE=docker compose -f $(COMPOSE_TOOLS_FILE)
COMPOSE_TOOLS_CMD_UP=$(COMPOSE_TOOLS_CMD_BASE) up --exit-code-from
COMPOSE_TOOLS_CMD_PULL=$(COMPOSE_TOOLS_CMD_BASE) pull

# COLORS
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)


TARGET_MAX_CHAR_NUM=20


define colored
	@echo '${GREEN}$1${RESET}'
endef

## Show help
help:
	${call colored, help is running...}
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  ${YELLOW}%-$(TARGET_MAX_CHAR_NUM)s${RESET} ${GREEN}%s${RESET}\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)



## Compile app
compile-instadiff-cli: vet
	./scripts/build/compile.sh
.PHONY: compile-instadiff-cli

## Build the app.
build: compile-instadiff-cli
.PHONY: build

## recreate all generated code and documentation.
codegen:
	$(COMPOSE_TOOLS_CMD_UP) go-generate go-generate
.PHONY: codegen

## recreate all generated code and swagger documentation and format code.
generate: codegen format-project vet
.PHONY: generate

## vet project
vet:
	./scripts/linting/run-vet.sh
.PHONY: vet

## Run full linting
lint-full:
	$(COMPOSE_TOOLS_CMD_UP) lint-full lint-full
.PHONY: lint-full

## Run linting for build pipeline
lint-pipeline:
	$(COMPOSE_TOOLS_CMD_UP) lint-pipeline lint-pipeline
.PHONY: lint-pipeline

## Run linting for sonar report
lint-sonar:
	$(COMPOSE_TOOLS_CMD_UP) lint-sonar lint-sonar
.PHONY: lint-sonar

## Test all packages
test:
	./scripts/tests/run.sh
.PHONY: test

test-regression: test
.PHONY: test-regression

## Test coverage report.
test-cover:
	./scripts/tests/coverage.sh
.PHONY: test-cover

prepare-cover-report: test-cover
	$(COMPOSE_TOOLS_CMD_UP) prepare-cover-report prepare-cover-report
.PHONY: prepare-cover-report

## Tests sonar report generate.
test-sonar-report:
	./scripts/tests/sonar-report.sh
.PHONY: test-sonar-report

## Installs tools from vendor.
install-tools:
	echo "Installing ${GOTOOLS_IMAGE_TAG}"
	$(COMPOSE_TOOLS_CMD_PULL)
.PHONY: install-tools

## Sync vendor of root project and tools.
sync-vendor:
	./scripts/sync-vendor.sh
.PHONY: sync-vendor

## Docker compose up
docker-up:
	${call colored, docker is running...}
	docker-compose -f ./docker-compose.yml up -d

.PHONY: docker-up

## Docker compose down
docker-down:
	${call colored, docker is running...}
	docker-compose -f ./docker-compose.yml down

.PHONY: docker-down

## Fix imports sorting.
imports:
	${call colored, fix-imports is running...}
	$(COMPOSE_TOOLS_CMD_UP) fix-imports fix-imports
.PHONY: imports

## Format code.
fmt:
	${call colored, fmt is running...}
	$(COMPOSE_TOOLS_CMD_UP) fix-fmt fix-fmt
.PHONY: fmt

## Format code and sort imports.
format-project: fmt imports
.PHONY: format-project

## Open coverage report.
open-cover-report: test-cover prepare-cover-report
	./scripts/open-coverage-report.sh
.PHONY: open-cover-report

## Update readme coverage badge.
update-readme-cover: test-cover
	$(COMPOSE_TOOLS_CMD_UP) update-readme-coverage update-readme-coverage
.PHONY: update-readme-cover

## Release
release:
	./scripts/release/release.sh
.PHONY: release

## Release local snapshot
release-local-snapshot:
	./scripts/release/local-snapshot-release.sh
.PHONY: release-local-snapshot

## Check goreleaser config.
check-releaser:
	./scripts/release/check.sh
.PHONY: check-releaser

## Issue new release.
new-version: test build
	./scripts/release/new-version.sh
.PHONY: new-release

.DEFAULT_GOAL := test
