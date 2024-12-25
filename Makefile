.DEFAULT_GOAL := help
GO := go
LDFLAGS := '-w -s'
BUILD_DIR := ./build
APP_BIN := $(BUILD_DIR)/namemyserver
APP_VERSION := 1.0

SOURCE_FILES := $(shell find . -type f -name "*.go")

COVERAGE_FILE := $(BUILD_DIR)/coverage.out

# source .env file if it exists, will make all variables available
# this include makes the help target not work, need to debug why
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: help
help: ## prints help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: install-tools
install-tools: ## installs tools required for development and building the project
	# TODO pin versions
	go install github.com/air-verse/air@latest
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2
	go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@v2.2.0
	# assumes ~/.local/bin is in $PATH
	curl -fsSL -o ~/.local/bin/dbmate https://github.com/amacneil/dbmate/releases/latest/download/dbmate-linux-amd64 && chmod +x ~/.local/bin/dbmate


$(APP_BIN): $(SOURCE_FILES)
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags $(LDFLAGS) -o $(APP_BIN) ./cmd/namemyserver

build: templ $(APP_BIN) ## builds namemyserver's binary for production use in the current machine's architecture

.PHONY: templ
templ: ## generates templ go code based on templ templates
	templ generate

.PHONY: test
test: ## runs all namemyserver tests
	@$(GO) test ./...

$(COVERAGE_FILE): $(SOURCE_FILES)
	@$(GO) test -coverprofile=$(COVERAGE_FILE) ./...

coverage: $(COVERAGE_FILE) ## tests namemyserver codebase and generates a coverage file

see-coverage: coverage ## showcases the coverage in a browser using html output
	@$(GO) tool cover -html=$(COVERAGE_FILE)

.PHONY: clean
clean: ## removes build assets generated by the build target from the system
	@rm -rf $(BUILD_DIR)/*

.PHONY: lint
lint: ## detects flaws in the code and checks for style
	@golangci-lint run

.PHONY: format
format: ## formats the codebase using golangci-lint linters
	@golangci-lint run --fix
	@templ fmt ./internal/templates

.PHONY: docker
docker: ## builds the application's docker image
	@docker build --progress=plain -t davidonium/namemyserver:$(APP_VERSION) .
	@docker tag davidonium/namemyserver:$(APP_VERSION) davidonium/namemyserver:latest


.PHONY: dbmigrate
dbmigrate:
	dbmate up

.PHONY: dbreset
dbreset:
	dbmate drop
	dbmate create
	dbmate up

.PHONY: dev
dev: ## launches the app with support for live reload of the server on file change
	# see https://github.com/air-verse/air
	air
