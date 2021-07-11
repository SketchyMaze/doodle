SHELL := /bin/bash

VERSION=$(shell egrep -e 'Version\s+=' pkg/branding/branding.go | head -n 1 | cut -d '"' -f 2)
BUILD=$(shell git describe --always)
BUILD_DATE=$(shell date +"%Y-%m-%dT%H:%M:%S%z")
CURDIR=$(shell curdir)

# Inject the build version (commit hash) into the executable.
LDFLAGS := -ldflags "-X main.Build=$(BUILD) -X main.BuildDate=$(BUILD_DATE)"
LDFLAGS_W := -ldflags "-X main.Build=$(BUILD) -X main.BuildDate=$(BUILD_DATE) -H windowsgui"

# `make setup` to set up a new environment, pull dependencies, etc.
.PHONY: setup
setup: clean
	go get -u git.kirsle.net/go/bindata/...
	go get ./...

# `make build` to build the binary.
.PHONY: build
build:
	go build $(LDFLAGS) -i -o bin/sketchymaze cmd/doodle/main.go
	go build $(LDFLAGS) -i -o bin/doodad cmd/doodad/main.go

# `make buildall` to run all build steps including doodads and bindata.
.PHONY: buildall
buildall: doodads bindata build

# `make build-free` to build the binary in free mode.
.PHONY: build-free
build-free:
	gofmt -w .
	go build $(LDFLAGS) -tags="shareware" -i -o bin/sketchymaze cmd/doodle/main.go
	go build $(LDFLAGS) -tags="shareware" -i -o bin/doodad cmd/doodad/main.go

# `make build-debug` to build the binary in developer mode.
.PHONY: build-debug
build-debug:
	gofmt -w .
	go build $(LDFLAGS) -tags="developer" -i -o bin/sketchymaze cmd/doodle/main.go
	go build $(LDFLAGS) -tags="developer" -i -o bin/doodad cmd/doodad/main.go

# `make bindata` generates the embedded binary assets package.
.PHONY: bindata
bindata:
	go-bindata -pkg bindata -o pkg/bindata/bindata.go assets/...

# `make bindata-dev` generates the debug version of bindata package.
.PHONY: bindata-dev
bindata-dev:
	go-bindata -debug -pkg bindata -o pkg/bindata/bindata.go assets/...

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
mingw: doodads bindata
	env CGO_ENABLED="1" CC="/usr/bin/x86_64-w64-mingw32-gcc" \
		GOOS="windows" CGO_LDFLAGS="-lmingw32 -lSDL2" CGO_CFLAGS="-D_REENTRANT" \
		go build $(LDFLAGS_W) -i -o bin/sketchymaze.exe cmd/doodle/main.go
	env CGO_ENABLED="1" CC="/usr/bin/x86_64-w64-mingw32-gcc" \
		GOOS="windows" CGO_LDFLAGS="-lmingw32 -lSDL2" CGO_CFLAGS="-D_REENTRANT" \
		go build $(LDFLAGS) -i -o bin/doodad.exe cmd/doodad/main.go

# `make mingw-free` for Windows binary in free mode.
.PHONY: mingw-free
mingw-free: doodads bindata
	env CGO_ENABLED="1" CC="/usr/bin/x86_64-w64-mingw32-gcc" \
		GOOS="windows" CGO_LDFLAGS="-lmingw32 -lSDL2" CGO_CFLAGS="-D_REENTRANT" \
		go build $(LDFLAGS_W) -tags="shareware" -i -o bin/sketchymaze.exe cmd/doodle/main.go
	env CGO_ENABLED="1" CC="/usr/bin/x86_64-w64-mingw32-gcc" \
		GOOS="windows" CGO_LDFLAGS="-lmingw32 -lSDL2" CGO_CFLAGS="-D_REENTRANT" \
		go build $(LDFLAGS) -tags="shareware" -i -o bin/doodad.exe cmd/doodad/main.go

# `make release` runs the release.sh script, must be run
# after `make dist`
.PHONY: release
release:
	./scripts/release.sh

# `make mingw-release` runs a FULL end-to-end release of Linux and Windows
# binaries of the game, zipped and tagged and ready to go.
.PHONY: mingw-release
mingw-release: doodads bindata build mingw __dist-common release

# `make osx` to cross-compile a Mac OS binary with osxcross.
# .PHONY: osx
# osx: doodads bindata
# 	CGO_ENABLED=1 CC=[path-to-osxcross]/target/bin/[arch]-apple-darwin[version]-clang GOOS=darwin GOARCH=[arch] go build -tags static -ldflags "-s -w" -a


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
dist: doodads bindata build __dist-common

# `make dist-free` builds and tars up a release in shareware mode.
.PHONY: dist-free
dist-free: doodads bindata build-free __dist-common

# Common logic behind `make dist`
.PHONY: __dist-common
__dist-common:
	mkdir -p dist/sketchymaze-$(VERSION)
	cp bin/* dist/sketchymaze-$(VERSION)/
	cp -r README.md Changes.md "Open Source Licenses.md" rtp dist/sketchymaze-$(VERSION)/
	if [[ -d ./guidebook ]]; then cp -r guidebook dist/sketchymaze-$(VERSION)/; fi
	rm -rf dist/sketchymaze-$(VERSION)/rtp/.git
	ln -sf sketchymaze-$(VERSION) dist/sketchymaze-latest
	cd dist && tar -czvf sketchymaze-$(VERSION).tar.gz sketchymaze-$(VERSION)
	cd dist && zip -r sketchymaze-$(VERSION).zip sketchymaze-$(VERSION)

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
