SHELL := /bin/bash

VERSION=$(shell grep -e 'Version =' doodle.go | head -n 1 | cut -d '"' -f 2)
BUILD=$(shell git describe --always)
CURDIR=$(shell curdir)

# Inject the build version (commit hash) into the executable.
LDFLAGS := -ldflags "-X main.Build=$(BUILD)"

# `make setup` to set up a new environment, pull dependencies, etc.
.PHONY: setup
setup: clean
	go get ./...

# `make build` to build the binary.
.PHONY: build
build:
	gofmt -w .
	go build $(LDFLAGS) -i -o bin/doodle cmd/doodle/main.go
	go build $(LDFLAGS) -i -o bin/doodad cmd/doodad/main.go

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

# `make dist` builds and tars up a release.
.PHONY: dist
dist: build
	mkdir -p dist/doodle-$(VERSION)
	cp bin/* dist/doodle-$(VERSION)/
	cp -r assets fonts README.md Changes.md dist/doodle-$(VERSION)/
	cd dist && tar -czvf doodle-$(VERSION).tar.gz doodle-$(VERSION)
	cd dist && zip -r doodle-$(VERSION).zip doodle-$(VERSION)

# `make docker` to run the Docker builds
.PHONY: docker docker.ubuntu docker.debian __docker.dist
docker.ubuntu:
	mkdir -p docker/ubuntu
	./docker/dist-ubuntu.sh
docker.debian:
	mkdir -p docker/debian
	./docker/dist-debian.sh
docker: docker.ubuntu docker.debian
__docker.dist: dist
	cp dist/*.tar.gz dist/*.zip /mnt/export/

# `make clean` cleans everything up.
.PHONY: clean
clean:
	rm -rf bin dist
