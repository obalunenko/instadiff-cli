NAME=instadiff-cli
BIN_DIR=./bin

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
compile:
	${call colored, compile is running...}
	./scripts/build/compile.sh
.PHONY: compile

## recreate all generated code.
generate:
	${call colored, generate is running...}
	./scripts/codegen/generate.sh
.PHONY: generate

## vet project
vet:
	./scripts/linting/run-vet.sh
.PHONY: vet

## Run full linting
lint-full:
	./scripts/linting/run-linters.sh
.PHONY: lint-full

## Run linting for build pipeline
lint-pipeline:
	./scripts/linting/golangci-pipeline.sh
.PHONY: lint-pipeline

## Run linting for sonar report
lint-sonar:
	./scripts/linting/golangci-sonar.sh
.PHONY: lint-sonar

## Test all packages
test:
	${call colored, test is running...}
	./scripts/tests/run.sh
.PHONY: test

## Test coverage
test-cover:
	${call colored, test-cover is running...}
	./scripts/tests/coverage.sh
.PHONY: test-cover

## Installs tools from vendor.
install-tools:
	./scripts/install/vendored-tools.sh
.PHONY: install-tools

## Sync vendor of root project and tools.
sync-vendor:
	./scripts/sync-vendor.sh
.PHONY: sync-vendor

## Docker compose up
docker-up:
	${call colored, docker is running...}
	docker-compose -f ./docker-compose.yml up

.PHONY: docker-up

## Docker compose down
docker-down:
	${call colored, docker is running...}
	docker-compose -f ./docker-compose.yml down --volumes

.PHONY: docker-down

## Fix imports sorting.
imports:
	${call colored, fix-imports is running...}
	./scripts/style/fix-imports.sh
.PHONY: imports

## Format code.
fmt:
	${call colored, fmt is running...}
	./scripts/style/fmt.sh
.PHONY: fmt

## Format code and sort imports.
format-project: fmt imports
.PHONY: format-project

## Open coverage report.
open-cover-report: test-cover
	./scripts/open-coverage-report.sh
.PHONY: open-cover-report

## Update readme coverage badge.
update-readme-cover: test-cover
	./scripts/update-readme-coverage.sh
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
new-version: vet test build
	./scripts/release/new-version.sh
.PHONY: new-release

.DEFAULT_GOAL := test
