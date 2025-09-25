.PHONY: build help
PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

NAME := zeit
PREFIX := github.com/mrusme/
PROJECT := $(PREFIX)$(NAME)
VERSION := $(shell git describe --tags 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --verify HEAD)
DATE := $(shell date)

all: build

help: ## print this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}'

build: ## build
	@echo "Building with the following parameters:"
	@echo "VERSION = $(VERSION)"
	@echo "COMMIT  = $(COMMIT)"
	@echo "DATE    = $(DATE)"
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "-s -w -X \"${PROJECT}/runtime.Version=${VERSION}\" -X \"${PROJECT}/runtime.Commit=${COMMIT}\" -X \"${PROJECT}/runtime.Date=${DATE}\"" -o $(PWD)/build/$(NAME)
