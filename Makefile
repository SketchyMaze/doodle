SHELL := /bin/bash

VERSION=$(shell grep -e 'Version' doodle.go | head -n 1 | cut -d '"' -f 2)
BUILD=$(shell git describe --always)
CURDIR=$(shell curdir)

# Inject the build version (commit hash) into the executable.
LDFLAGS := -ldflags "-X main.Build=$(BUILD)"

# `make setup` to set up a new environment, pull dependencies, etc.
.PHONY: setup
setup: clean
	go get -u ./...

# `make build` to build the binary.
.PHONY: build
build:
	gofmt -w .
	go build $(LDFLAGS) -i -o bin/doodle cmd/doodle/main.go

# `make run` to run it in debug mode.
.PHONY: run
run:
	go run cmd/doodle/main.go -debug

# `make guitest` to run it in guitest mode.
.PHONY: guitest
guitest:
	go run cmd/doodle/main.go -debug -guitest

# `make test` to run unit tests.
.PHONY: test
test:
	go test ./...

# `make clean` cleans everything up.
.PHONY: clean
clean:
	rm -rf bin dist
