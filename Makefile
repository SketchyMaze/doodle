SHELL := /bin/bash

VERSION=$(shell grep -e 'Version =' pkg/branding/branding.go | head -n 1 | cut -d '"' -f 2)
BUILD=$(shell git describe --always)
BUILD_DATE=$(shell date -Iseconds)
CURDIR=$(shell curdir)

# Inject the build version (commit hash) into the executable.
LDFLAGS := -ldflags "-X main.Build=$(BUILD) -X main.BuildDate=$(BUILD_DATE)"

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

# `make build-free` to build the binary in free mode.
.PHONY: build-free
build-free:
	gofmt -w .
	go build $(LDFLAGS) -tags="shareware" -i -o bin/doodle cmd/doodle/main.go
	go build $(LDFLAGS) -tags="shareware" -i -o bin/doodad cmd/doodad/main.go

# `make build-debug` to build the binary in developer mode.
.PHONY: build-debug
build-debug:
	gofmt -w .
	go build $(LDFLAGS) -tags="developer" -i -o bin/doodle cmd/doodle/main.go
	go build $(LDFLAGS) -tags="developer" -i -o bin/doodad cmd/doodad/main.go

# `make wasm` builds the WebAssembly port.
.PHONY: wasm
wasm:
	cd wasm && make

# `make wasm-serve` builds and launches the WebAssembly server.
.PHONY: wasm-serve
wasm-serve: wasm
	sh -c 'sleep 1; xdg-open http://localhost:8080/' &
	cd wasm && go run server.go

# `make install` to install the Go binaries to your GOPATH.
.PHONY: install
install:
	go install git.kirsle.net/apps/doodle/cmd/...

# `make doodads` to build the doodads from the dev-assets folder.
.PHONY: doodads
doodads:
	cd dev-assets/doodads && ./build.sh

# `make mingw` to cross-compile a Windows binary with mingw.
.PHONY: mingw
mingw:
	env CGO_ENABLED="1" CC="/usr/bin/x86_64-w64-mingw32-gcc" \
		GOOS="windows" CGO_LDFLAGS="-lmingw32 -lSDL2" CGO_CFLAGS="-D_REENTRANT" \
		go build $(LDFLAGS) -i -o bin/doodle.exe cmd/doodle/main.go
		env CGO_ENABLED="1" CC="/usr/bin/x86_64-w64-mingw32-gcc" \
			GOOS="windows" CGO_LDFLAGS="-lmingw32 -lSDL2" CGO_CFLAGS="-D_REENTRANT" \
			go build $(LDFLAGS) -i -o bin/doodad.exe cmd/doodad/main.go



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
	cp -r assets fonts README.md dist/doodle-$(VERSION)/
	cd dist && tar -czvf doodle-$(VERSION).tar.gz doodle-$(VERSION)
	cd dist && zip -r doodle-$(VERSION).zip doodle-$(VERSION)

# `make docker` to run the Docker builds
.PHONY: docker docker.ubuntu docker.debian docker.fedora __docker.dist
docker.ubuntu:
	mkdir -p docker/ubuntu
	./docker/dist-ubuntu.sh
docker.debian:
	mkdir -p docker/debian
	./docker/dist-debian.sh
docker.fedora:
	mkdir -p docker/fedora
	./docker/dist-fedora.sh
docker: docker.ubuntu docker.debian docker.fedora
__docker.dist: dist
	cp dist/*.tar.gz dist/*.zip /mnt/export/

# `make clean` cleans everything up.
.PHONY: clean
clean:
	rm -rf bin dist docker/ubuntu docker/debian docker/fedora
