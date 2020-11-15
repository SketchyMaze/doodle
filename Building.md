# Building Doodle

* [Linux](#linux)
* [Windows Cross-Compile from Linux](#windows-cross-compile-from-linux)
* [Mac OS](#mac os)

## Complete App Build

You'll need the following git repositories:

* git.kirsle.net/apps/doodle - the game engine.
* git.kirsle.net/apps/doodle-masters - where built-in level files are kept.
* git.kirsle.net/apps/doodle-vendor - vendored libraries for Windows (SDL2.dll etc.)

```bash
# Copy fonts and levels in
$ cp /git/doodle-masters/levels assets/levels
$ cp /git/doodle-vendor/fonts assets/fonts

# Ensure you have bindata CLI command. NOTE: below repo is
# my fork of go-bindata, can find it elsewhere instead.
$ go get -u git.kirsle.net/go/bindata/...

# From the doodle repo
$ make bindata-dev  # TODO: populates the bindata .go modules.
$ go get ./...      # install dependencies etc.
```

The `make setup` command tries to do the above.

`make build` produces a local binary in the bin/ folder and `make dist`
will build an app for distribution in the dist/ folder.

Levels should be copied in from the doodle-masters repo into the
assets/levels/ folder before running `make bindata` to release the game.

## Fonts

The `fonts/` folder is git-ignored. The app currently uses font files here
named:

* `DejaVuSans.ttf` for sans-serif font.
* `DejaVuSans-Bold.ttf` for bold sans-serif font.
* `DejaVuSansMono.ttf` for monospace font.

These are the open source **DejaVu Sans [Mono]** fonts, so copy them in from
your `/usr/share/fonts/dejavu` folder or provide alternative fonts.

```bash
mkdir fonts
cp /usr/share/fonts/dejavu/{DejaVuSans.ttf,DejaVuSans-Bold.ttf,DejaVuSansMono.ttf} fonts/
```

## Makefile

Makefile commands for Linux:

* `make setup`: install Go dependencies and set up the build environment
* `make doodads`: build the default Doodads from sources in `dev-assets/`
* `make bindata`: embed the default doodads, levels and other assets into the
  Go program. `make bindata-dev` for lightweight dev versions that will read
  from the filesystem at runtime instead.
* `make build`: build the Doodle and Doodad binaries to the `bin/` folder.
* `make buildall`: runs all build steps: doodads, bindata, build.
* `make build-free`: build the shareware binaries to the `bin/` folder. See
  Build Tags below.
* `make build-debug`: build a debug binary (not release-mode) to the `bin/`
  folder. See Build Tags below.
* `make wasm`: build the WebAssembly output
* `make wasm-serve`: build the WASM output and then run the server.
* `make run`: run a local dev build of Doodle in debug mode
* `make guitest`: run a local dev build in the GUITest scene
* `make test`: run the test suite
* `make dist`: produce a zipped release tarball and zip file for your current
  environment and output into the `dist/` folder.
* `make docker`: run all the Dockerfiles from the `docker/` folder to produce
  dist builds for Debian, Fedora and Ubuntu. You may also run these builds
  individually:
  * `make docker.ubuntu`
  * `make docker.debian`
  * `make docker.fedora`
* `make clean`: clean all build artifacts

## Build Tags

### shareware

> Files ending with `_free.go` are for the shareware release as opposed to
> `_paid.go` for the full version.

Builds the game in the free shareware release mode.

Run `make build-free` to build the shareware binary.

Shareware releases of the game have the following changes compared to the default
(release) mode:

* No access to the Doodad Editor scene in-game (soft toggle)

### developer

> Files ending with `_developer.go` are for the developer build as opposed to
> `_release.go` for the public version.

Developer builds support extra features over the standard release version:

* Ability to write the JSON file format for Levels and Doodads.

Run `make build-debug` to build a developer version of the program.

## Linux

Dependencies are Go, SDL2 and SDL2_ttf:

```bash
# Fedora
sudo dnf -y install golang SDL2-devel SDL2_ttf-devel SDL2_mixer-devel

# Ubuntu and Debian
sudo apt -y install golang libsdl2-dev libsdl2-ttf-dev libsdl2-mixer-devel
```

## Mac OS

```bash
brew install golang sdl2 sdl2_ttf pkg-config
```

## Windows Cross-Compile from Linux

Install the Mingw C compiler:

```bash
# Fedora
sudo dnf -y install mingw64-gcc  # for 64-bit
sudo dnf -y install mingw32-gcc  # for 32-bit

# Arch Linux
pacman -S mingw-w64
```

Download the SDL2 Mingw development libraries [here](https://libsdl.org/download-2.0.php)
and SDL2_TTF from [here](https://www.libsdl.org/projects/).

Extract each and copy their library folder into the mingw path.

```bash
# e.g. /usr/x86_64-w64-mingw32 is usually the correct path, verify on e.g.
# Fedora with `rpm -ql mingw64-filesystem`
tar -xzvf SDL2_ttf-devel-2.0.15-mingw.tar.gz
cd SDL_ttf-2.0.15
sudo cp -r x86_64-w64-mingw32 /usr
```

Make and set permissions for Go to download the standard library for Windows:

```bash
mkdir /usr/lib/golang/pkg/windows_amd64
chown your_username /usr/lib/golang/pkg/windows_amd64
```

And run `make mingw` to build the Windows binary.

### Windows DLLs

For the .exe to run it will need SDL2.dll and such.

```
# SDL2.dll and SDL2_ttf.dll
cp /usr/x86_64-w64-mingw32/bin/SDL*.dll bin/
```

SDL2_ttf requires libfreetype, you can get its DLL here:
https://github.com/ubawurinna/freetype-windows-binaries
