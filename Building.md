# Building Doodle

* [Quickstart](#quickstart-with-bootstrap-py)
* [Detailed Instructions](#detailed-instructions)
* [Linux](#linux)
* [Flatpak for Linux](#flatpak-for-linux)
* [Windows Cross-Compile from Linux](#windows-cross-compile-from-linux)
* [Old Docs](#old-docs)

# Automated Release Scripts

For the quickest ways to fully end-to-end build Sketchy Maze for various
platforms to produce public release artifacts, see the following repos:

* [SketchyMaze/docker](https://git.kirsle.net/SketchyMaze/docker) provides a Dockerfile
  that fully end-to-end releases the latest version of the game for Linux and Windows:
  * Windows: .zip file
  * Linux: .tar.gz, .rpm, .deb
* [flatpak](https://code.sketchymaze.com/game/flatpak) is a Flatpak manifest for
  Linux distributions.

The Docker container depends on all the git servers being up, but if you have
the uber blob source code you can read the Dockerfile to see what it does.

# Quickstart with bootstrap.py

From any Unix-like system (Fedora, Ubuntu, macOS) the bootstrap.py script
is intended to set this app up, from scratch, _fast._ The basic steps are:

```bash
# Example from an Ubuntu 20.04 LTS fresh install running in a
# libvirt-manager virtual machine on a Fedora host.

# 1. Ensure your SSH keys have git clone permission for git.kirsle.net.
# For example just scp my keys from the host Fedora machine.
$ scp -r kirsle@192.168.122.1:.ssh/id_rsa* ~/.ssh/

# 2. git clone the Project Doodle repository.
$ git clone git@git.kirsle.net:apps/doodle
$ cd ./doodle

# 3. Run the bootstrap script.
$ python3 bootstrap.py
```

The bootstrap script will take care of the rest:

* `apt install` all the dependencies (golang, SDL2, etc.)
* `git clone` various other repositories into a "deps/" folder in doodle's
  directory. These are things like my Go render library `go/render` and
  `go/ui` as well as the doodle-rtp runtime package (sound effects, etc.)
  all of which are hosted on git.kirsle.net.
* Build and install the `doodad` tool so it can generate the builtin
  doodads, and build and release a full distribution of the game.

It should work on Fedora-likes, Debian-likes and macOS all the same.
It even runs on the Pine64 Pinephone (ARM64) with Mobian!

MacOS is expected to have [homebrew](https://brew.sh) installed.
MP3 support issues? [See here](https://github.com/veandco/go-sdl2/issues/299#issuecomment-611681191).

**To do:** the most important repositories, like the game itself, are
also mirrored on GitHub. Other supporting repos need mirroring too, or
otherwise, full source tarballs (the result of bootstrap.py) will be
built and archived somewhere safe for posterity in case git.kirsle.net
ever goes away. The doodle mirror is at <https://github.com/kirsle/doodle>
(private repository) and the others are there too (go/render, go/ui, etc.)

# Detailed Instructions

For building the app the hard way, and in-depth instructions, read
this section. You'll need the following git repositories:

* `git.kirsle.net/SketchyMaze/doodle` - the game engine.
* `git.kirsle.net/apps/SketchyMaze/masters` - where built-in level files are kept,
  as well as master GIMP drawings for sprites and such but only the levels
  are necessary to build the app properly. Tho even then the app would
  work fine with no levels built in!
* `git.kirsle.net/apps/SketchyMaze/vendor` - vendored libraries for Windows (SDL2.dll etc.)
* `git.kirsle.net/apps/SketchyMaze/rtp` - runtime package (sounds and music mostly)

The [docker](https://git.kirsle.net/SketchyMaze/docker) repo will
be more up-to-date than the instructions below, as that repo actually has
runnable code in the Dockerfile!

```bash
# Clone all the repos down to your project folder
git clone git@git.kirsle.net:apps/SketchyMaze/rtp rtp
git clone git@git.kirsle.net:apps/SketchyMaze/vendor vendor
git clone git@git.kirsle.net:apps/SketchyMaze/masters masters
git clone git@git.kirsle.net:apps/doodle doodle

# Enter doodle/ project
cd doodle/

# Copy fonts and levels in
cp ../masters/levels assets/levels
cp ../vendor/fonts assets/fonts
mkdir rtp && cp -r ../rtp/* rtp/

# From the doodle repo:
make setup  # -or-
go get ./...      # install dependencies etc.

# The app should build now. Build and install the doodad tool.
go install git.kirsle.net/SketchyMaze/doodle/cmd/doodad
doodad --version
# "doodad version 0.3.0-alpha build ..."

# Build and release the game into the dist/ folder.
# This will: generate builtin doodads, bundle them with bindata,
# and create a tarball in the dist/ folder.
make dist

# Build a cross-compiled Windows target from Linux.
# (you'd run before `make dist` to make an uber release)
make mingw

# After make dist, `make release` will carve up Linux
# and Windows (mingw) builds and zip them up nicely.
make release
```

The `make setup` command tries to do the above.

`make build` produces a local binary in the bin/ folder and `make dist`
will build an app for distribution in the dist/ folder.

Levels should be copied in from the doodle-masters repo into the
assets/levels/ folder before building the game.

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

The doodle-vendor repo has copies of these fonts.

## Makefile

Makefile commands for Unix-likes:

* `make setup`: install Go dependencies and set up the build environment
* `make doodads`: build the default Doodads from sources in `deps/doodads/`
* `make build`: build the Doodle and Doodad binaries to the `bin/` folder.
* `make buildall`: runs all build steps: doodads, build.
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

# Dependencies

The bootstrap.py script lists dependencies for Fedora, Debian and macOS.
Also here for clarity, hopefully not out-of-date:

```bash
# Fedora-likes
sudo dnf install make golang SDL2-devel SDL2_ttf-devel \
     SDL2_mixer-devel

# Debian and Ubuntu
sudo dnf install make golang libsdl2-dev libsdl2-ttf-dev \
     libsdl2-mixer-dev

# macOS via Homebrew (https://brew.sh)
brew install golang sdl2 sdl2_ttf sdl2_mixer pkg-config
```

## Flatpak for Linux

The repo for this is at <https://code.sketchymaze.com/game/flatpak>.

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

# Old Docs

## Build Tags

These aren't really used much anymore but documented here:

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
