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
	go get ./...

# `make build` to build the binary.
.PHONY: build
build:
	go build $(LDFLAGS) -o bin/sketchymaze cmd/doodle/main.go
	go build $(LDFLAGS) -o bin/doodad cmd/doodad/main.go

# `make buildall` to run all build steps including doodads.
.PHONY: buildall
buildall: doodads build

# `make build-free` to build the binary in free mode.
.PHONY: build-free
build-free:
	gofmt -w .
	go build $(LDFLAGS) -tags="shareware" -o bin/sketchymaze cmd/doodle/main.go
	go build $(LDFLAGS) -tags="shareware" -o bin/doodad cmd/doodad/main.go

# `make build-debug` to build the binary in developer mode.
.PHONY: build-debug
build-debug:
	gofmt -w .
	go build $(LDFLAGS) -tags="developer" -o bin/sketchymaze cmd/doodle/main.go
	go build $(LDFLAGS) -tags="developer" -o bin/doodad cmd/doodad/main.go

# `make bindata` generates the embedded binary assets package.
.PHONY: bindata
bindata:
	echo "make bindata: deprecated in favor of Go 1.16 embed; nothing was done"

# `make bindata-dev` generates the debug version of bindata package.
.PHONY: bindata-dev
bindata-dev:
	echo "make bindata-dev: deprecated in favor of Go 1.16 embed; nothing was done"

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
	go install git.kirsle.net/SketchyMaze/doodle/cmd/...

# `make doodads` to build the doodads from the deps/doodads folder.
.PHONY: doodads
doodads:
	cd deps/doodads && ./build.sh > /dev/null

# `make mingw` to cross-compile a Windows binary with mingw.
.PHONY: mingw
mingw:
	env CGO_ENABLED="1" CC="/usr/bin/x86_64-w64-mingw32-gcc" \
		GOOS="windows" CGO_LDFLAGS="-lmingw32 -lSDL2" CGO_CFLAGS="-D_REENTRANT" \
		go build $(LDFLAGS_W) -i -o bin/sketchymaze.exe cmd/doodle/main.go
	env CGO_ENABLED="1" CC="/usr/bin/x86_64-w64-mingw32-gcc" \
		GOOS="windows" CGO_LDFLAGS="-lmingw32 -lSDL2" CGO_CFLAGS="-D_REENTRANT" \
		go build $(LDFLAGS) -i -o bin/doodad.exe cmd/doodad/main.go

# `make mingw32` to cross-compile a Windows binary with mingw (32-bit).
.PHONY: mingw32
mingw32:
	env CGO_ENABLED="1" CC="/usr/bin/i686-w64-mingw32-gcc" \
		GOOS="windows" CGO_LDFLAGS="-lmingw32 -lSDL2" CGO_CFLAGS="-D_REENTRANT" \
		go build $(LDFLAGS_W) -i -o bin/sketchymaze.exe cmd/doodle/main.go
	env CGO_ENABLED="1" CC="/usr/bin/i686-w64-mingw32-gcc" \
		GOOS="windows" CGO_LDFLAGS="-lmingw32 -lSDL2" CGO_CFLAGS="-D_REENTRANT" \
		go build $(LDFLAGS) -i -o bin/doodad.exe cmd/doodad/main.go

# `make mingw-free` for Windows binary in free mode.
.PHONY: mingw-free
mingw-free:
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

# `make release32` runs release with ARCH_LABEL=32bit to product
# artifacts targeting an i386 architecture (e.g. in rpm and deb packages
# metadata about the release)
.PHONY: release32
release32:
	env ARCH_LABEL=32bit ./scripts/release.sh

# `make appimage` builds an AppImage, run it after `make dist`
.PHONY: appimage
appimage:
	./appimage.sh

# `make mingw-release` runs a FULL end-to-end release of Linux and Windows
# binaries of the game, zipped and tagged and ready to go.
.PHONY: mingw-release
mingw-release: doodads build mingw __dist-common release

.PHONY: mingw32-release
mingw32-release: doodads build mingw32 __dist-common release32

# `make from-docker64` is an internal command run by the Dockerfile to build the
# game - assumes doodads and assets are in the right spot already.
.PHONY: from-docker64
.PHONY: from-docker32
from-docker64: build mingw __dist-common
	ARCH=x86_64 make appimage
	make release
from-docker32: build mingw32 __dist-common
	ARCH=i686 make appimage
	make release32

# `make osx` to cross-compile a Mac OS binary with osxcross.
# .PHONY: osx
# osx: doodads
# 	CGO_ENABLED=1 CC=[path-to-osxcross]/target/bin/[arch]-apple-darwin[version]-clang GOOS=darwin GOARCH=[arch] go build -tags static -ldflags "-s -w" -a


# `make run` to run it from source.
.PHONY: run
run:
	go run cmd/doodle/main.go

# `make debug` to run it in -debug mode.
.PHONY: debug
debug:
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
dist: doodads build __dist-common

# `make docker` runs the Dockerfile to do a full release for 64-bit and 32-bit Linux
# and Windows apps.
.PHONY: docker
docker:
	./scripts/docker-build.sh

# `make dist-free` builds and tars up a release in shareware mode.
.PHONY: dist-free
dist-free: doodads build-free __dist-common

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

# `make clean` cleans everything up.
.PHONY: clean
clean:
	rm -rf bin dist docker-artifacts
