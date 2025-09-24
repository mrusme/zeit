.PHONY: build help
PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

VERSION ?= $(shell git describe --tags)
COMMIT ?= $(shell git rev-parse --verify HEAD)
DATE ?= $(shell date)

all: build

help: ## print this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}'

build: ## build
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "-s -w -X 'runtime.version=${VERSION}' -X 'runtime.commit=${COMMIT}' -X 'runtime.date=${DATE}' -X 'main.builtBy=Makefile'" -o $(PWD)/build/0C

