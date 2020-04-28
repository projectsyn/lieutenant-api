# Project parameters
BINARY_NAME ?= lieutenant-api

VERSION ?= $(shell git describe --tags --always --dirty --match=v* || (echo "command failed $$?"; exit 1))

IMAGE_NAME ?= docker.io/projectsyn/$(BINARY_NAME):$(VERSION)

# Antora variables
docker_cmd  ?= docker
docker_opts ?= --rm --tty --user "$$(id -u)"

antora_cmd  ?= $(docker_cmd) run $(docker_opts) --volume "$${PWD}":/antora vshn/antora:2.3.0
antora_opts ?= --cache-dir=.cache/antora

vale_img ?= docker.io/vshn/vale:2.1.1
openapi_generator_img ?= docker.io/openapitools/openapi-generator:cli-v4.3.0

vale_cmd ?= $(docker_cmd) run $(docker_opts) --volume "$${PWD}"/docs/modules/ROOT/pages:/pages $(vale_img) \
	--minAlertLevel=error \
	--config=/pages/.vale.ini /pages

openapi_validate_cmd ?= $(docker_cmd) run $(docker_opts) --volume "$${PWD}"/openapi.yaml:/openapi.yaml $(openapi_generator_img) \
	validate -i /openapi.yaml

openapi_generate_docs_cmd ?= $(docker_cmd) run $(docker_opts) --volume "$${PWD}":/local $(openapi_generator_img) \
	generate -i /local/openapi.yaml \
	--generator-name asciidoc \
	--output /local/docs/modules/ROOT/pages

# Go parameters
GOCMD   ?= go
GOBUILD ?= $(GOCMD) build
GOCLEAN ?= $(GOCMD) clean
GOTEST  ?= $(GOCMD) test
GOGET   ?= $(GOCMD) get

.PHONY: all
all: test build

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
	$(antora_cmd) $(antora_opts) $<

.PHONY: check
check: generate-api-docs
	$(vale_cmd)
