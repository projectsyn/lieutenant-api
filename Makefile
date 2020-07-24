MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := all
.DELETE_ON_ERROR:
.SUFFIXES:

DOCKER_CMD   ?= docker
DOCKER_ARGS  ?= --rm --user "$$(id -u)" --volume "$${PWD}:/src" --workdir /src

# Project parameters
BINARY_NAME ?= lieutenant-api

VERSION ?= $(shell git describe --tags --always --dirty --match=v* || (echo "command failed $$?"; exit 1))

IMAGE_NAME ?= docker.io/projectsyn/$(BINARY_NAME):$(VERSION)

# Antora variables
ANTORA_CMD  ?= $(DOCKER_CMD) run $(DOCKER_ARGS) --volume "$${PWD}":/antora vshn/antora:2.3.0
ANTORA_ARGS ?= --cache-dir=.cache/antora


VALE_CMD  ?= $(DOCKER_CMD) run $(DOCKER_ARGS) --volume "$${PWD}"/docs/modules:/pages vshn/vale:2.1.1
VALE_ARGS ?= --minAlertLevel=error --config=/pages/ROOT/pages/.vale.ini /pages

openapi_generator_img ?= docker.io/openapitools/openapi-generator:cli-v4.3.0
openapi_validate_cmd ?= $(DOCKER_CMD) run $(DOCKER_ARGS) --volume "$${PWD}"/openapi.yaml:/openapi.yaml $(openapi_generator_img) \
	validate -i /openapi.yaml

openapi_generate_docs_cmd ?= $(DOCKER_CMD) run $(DOCKER_ARGS) --volume "$${PWD}":/local $(openapi_generator_img) \
	generate -i /local/openapi.yaml \
	--generator-name asciidoc \
	--output /local/docs/modules/ROOT/pages

# Linting parameters
YAML_FILES      ?= $(shell find . -type f -name '*.yaml' -or -name '*.yml')
YAMLLINT_ARGS   ?= --no-warnings
YAMLLINT_CONFIG ?= .yamllint.yml
YAMLLINT_IMAGE  ?= docker.io/cytopia/yamllint:latest
YAMLLINT_DOCKER ?= $(DOCKER_CMD) run $(DOCKER_ARGS) $(YAMLLINT_IMAGE)

# Go parameters
GOCMD   ?= go
GOBUILD ?= $(GOCMD) build
GOCLEAN ?= $(GOCMD) clean
GOTEST  ?= $(GOCMD) test
GOGET   ?= $(GOCMD) get

.PHONY: all
all: lint test build

.PHONY: validate
validate:
	$(openapi_validate_cmd)

.PHONY: generate
generate:
	go generate main.go

.PHONY: build
build: generate
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -v \
		-o $(BINARY_NAME) \
		-ldflags "-X main.Version=$(VERSION) -X 'main.BuildDate=$(shell date)'" \
		main.go
	@echo built '$(VERSION)'

.PHONY: test
test: generate
	$(GOTEST) -v -cover ./...

.PHONY: run
run: generate
	go run main.go

.PHONY: watch
watch: generate
	go run github.com/cosmtrek/air

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

.PHONY: docker
docker:
	DOCKER_BUILDKIT=1 docker build -t $(IMAGE_NAME) .
	@echo built image $(IMAGE_NAME)

.PHONY: generate-api-docs
generate-api-docs:
	$(openapi_generate_docs_cmd)

.PHONY: docs
docs: generate-api-docs $(web_dir)/index.html

$(web_dir)/index.html: playbook.yml $(pages)
	$(ANTORA_CMD) $(ANTORA_ARGS) $<


.PHONY: lint
lint: lint_yaml lint_adoc

.PHONY: lint_yaml
lint_yaml: $(YAML_FILES)
	$(YAMLLINT_DOCKER) -f parsable -c $(YAMLLINT_CONFIG) $(YAMLLINT_ARGS) -- $?

.PHONY: lint_adoc
lint_adoc:
	$(VALE_CMD) $(VALE_ARGS)
