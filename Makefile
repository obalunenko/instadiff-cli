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



## Cross system compile
compile:
	${call colored, compile is running...}
	./scripts/cross-compile.sh
.PHONY: compile

## lint project
lint:
	${call colored, lint is running...}
	./scripts/run-linters.sh
.PHONY: lint

## format markdown files in project
pretty-markdown:
	find . -name '*.md' -not -wholename './vendor/*' | xargs prettier --write
.PHONY: pretty-markdown

## Test all packages
test:
	${call colored, test is running...}
	./scripts/run-tests.sh
.PHONY: test

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

## Fix imports sorting
imports:
	${call colored, sort and group imports...}
	./scripts/fix-imports.sh
.PHONY: imports

## dependencies - fetch all dependencies for sripts
dependencies:
	${call colored, dependensies is running...}
	./scripts/get-dependencies.sh
.PHONY dependencies



.DEFAULT_GOAL := test

