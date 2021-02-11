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
	./scripts/compile.sh
.PHONY: compile

## recreate all generated code.
generate:
	${call colored, generate is running...}
	./scripts/generate.sh
.PHONY: generate

## lint project
lint:
	${call colored, lint is running...}
	./scripts/run-linters.sh
.PHONY: lint

lint-ci:
	${call colored, lint_ci is running...}
	./scripts/run-linters-ci.sh
.PHONY: lint-ci

## format markdown files in project
pretty-markdown:
	find . -name '*.md' -not -wholename './vendor/*' | xargs prettier --write
.PHONY: pretty-markdown

## Test all packages
test:
	${call colored, test is running...}
	./scripts/run-tests.sh
.PHONY: test

test-ci: test
	${call colored, test-ci is running...}
	./scripts/run-tests-ci.sh
.PHONY: test-ci

## Test coverage
test-cover:
	${call colored, test-cover is running...}
	./scripts/coverage.sh
.PHONY: test-cover

new-version: lint test compile
	${call colored, new version is running...}
	./scripts/version.sh
.PHONY: new-version

## Release
release:
	${call colored, release is running...}
	./scripts/release.sh
.PHONY: release


## Installs tools from vendor.
install-tools:
	./scripts/install-tools.sh
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
	./scripts/fix-imports.sh
.PHONY: imports

## Format code.
fmt:
	${call colored, fmt is running...}
	./scripts/fmt.sh
.PHONY: fmt

## Format code and sort imports.
format-project: fmt imports
.PHONY: format-project

## vet project
vet:
	${call colored, vet is running...}
	./scripts/vet.sh
.PHONY: vet

.DEFAULT_GOAL := test
