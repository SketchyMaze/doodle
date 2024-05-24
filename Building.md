# Building Doodle

- [Building Doodle](#building-doodle)
- [Dockerfile](#dockerfile)
- [Automated Release Scripts](#automated-release-scripts)
- [Go Environment](#go-environment)
- [Quickstart with bootstrap.py](#quickstart-with-bootstrappy)
- [Detailed Instructions](#detailed-instructions)
  - [Fonts](#fonts)
  - [Makefile](#makefile)
- [Dependencies](#dependencies)
  - [Flatpak for Linux](#flatpak-for-linux)
  - [Windows Cross-Compile from Linux](#windows-cross-compile-from-linux)
    - [Windows DLLs](#windows-dlls)
  - [Build on macOS from scratch](#build-on-macos-from-scratch)
- [WebAssembly](#webassembly)
- [Build Tags](#build-tags)
  - [doodad](#doodad)
  - [dpp](#dpp)

# Dockerfile

The Dockerfile in this git repo may be the quickest way to fully
release the game for as many platforms as possible. Run it from a
64-bit host Linux system and it will generate Linux and Windows
releases for 64-bit and 32-bit Intel CPUs.

It depends on your git clone of doodle to be fully initialized
(e.g., you have run the bootstrap.py script and a `make dist`
would build a release for your current system, with doodads and
runtime assets all in the right places).

Run `make docker` and the results will be in the
`artifacts/release` folder in your current working directory.

**Fedora notes (SELinux):** if you run this from a Fedora host
you will need to `sudo setenforce permissive` to allow the
Dockerfile to mount the artifacts/release folder to export its
results.

# Automated Release Scripts

Other Dockerfiles and scripts used to release the game:

* [SketchyMaze/docker](https://git.kirsle.net/SketchyMaze/docker) provides a Dockerfile
  that fully end-to-end releases the latest version of the game for Linux and Windows. 64bit and 32bit versions that freshly clone the
  game from git and output their respective CPU release artifacts:
  * Windows: .zip file
  * Linux: .tar.gz, .rpm, .deb
* [flatpak](https://code.sketchymaze.com/game/flatpak) is a Flatpak manifest for
  Linux distributions.

The Docker container depends on all the git servers being up, but if you have
the uber blob source code you can read the Dockerfile to see what it does.

# Go Environment

Part of the build scripts involve building and running the `doodad` command
from this repo in order to generate the game's built-in doodads. For this to
work smoothly from your Linux or macOS build environment, you may need to
ensure that your `${GOPATH}/bin` directory is on your `$PATH` by, for example,
configuring this in your bash/zsh profile:

```bash
export GOPATH="${HOME}/go"
export PATH="${PATH}:${GOPATH}/bin"
```

For a complete example, see the "Build on macOS from scratch" section below.

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
ever goes away. The doodle mirror is at <https://github.com/SketchyMaze/doodle>
(private repository) and the others are there too (go/render, go/ui, etc.)

# Detailed Instructions

For building the app the hard way, and in-depth instructions, read
this section. You'll need the following git repositories:

* `git.kirsle.net/SketchyMaze/doodle` - the game engine.
* `git.kirsle.net/SketchyMaze/assets` - where built-in level files are kept (optional)
* `git.kirsle.net/SketchyMaze/vendor` - vendored libraries for Windows (SDL2.dll etc.)
* `git.kirsle.net/SketchyMaze/rtp` - runtime package (sounds and music mostly)
* `git.kirsle.net/SketchyMaze/doodads` - sources to compile the built-in doodads.

The [docker](https://git.kirsle.net/SketchyMaze/docker) repo will
be more up-to-date than the instructions below, as that repo actually has
runnable code in the Dockerfile!

```bash
# Clone all the repos down to your project folder
git clone https://git.kirsle.net/SketchyMaze/rtp rtp
git clone https://git.kirsle.net/SketchyMaze/vendor vendor
git clone https://git.kirsle.net/SketchyMaze/masters masters
git clone https://git.kirsle.net/SketchyMaze/doodle doodle
git clone https://git.kirsle.net/SketchyMaze/doodads doodle/deps/doodads

# Enter doodle/ project
cd doodle/

# Copy fonts and levels in
cp ../assets/levelpacks assets/levelpacks
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

`make build` produces a local binary in the bin/ folder and `make dist`
will build an app for distribution in the dist/ folder.

The bootstrap.py script does all of the above up to `make dist` so if you need
fully release the game by hand (e.g. on a macOS host) you can basically get away
with:

1. Clone the doodle repo and cd into it
2. Run `bootstrap.py` to fully set up your OS with dependencies and build a
   release quality version of the game with all latest assets (the script finishes
   with a `make dist`).
3. Run `make release` to package the dist/ artifact into platform specific
   release artifacts (.rpm/.deb/.tar.gz bundles for Linux, .zip for Windows,
   .dmg if running on macOS) which output into the dist/release/ folder.

Before step 3 you may want to download the latest Guidebook to bundle with
the game (optional). Grab and extract the tarball and run `make dist && make release`:

```bash
wget -O - https://download.sketchymaze.com/guidebook.tar.gz | tar -xzvf -
```

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

The repo for this is at <https://git.kirsle.net/SketchyMaze/flatpak>.

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

## Build on macOS from scratch

Here are some detailed instructions how to build Sketchy Maze from a fresh
install of macOS Ventura that assumes no previous software or configuration
was applied to the system yet.

Install homebrew: https://brew.sh pay attention to the instructions at the end
of the install to set up your zsh profile for homebrew to work correctly.

Clone the doodle repository:

```bash
git clone https://git.kirsle.net/SketchyMaze/doodle
cd doodle
```

Note: on a fresh install, invoking the `git` command may cause macOS to install
developer tools and Xcode. After installed, run the git clone again to finish
cloning the repository.

Set your Go environment variables: edit your ~/.zprofile and ensure that $GOPATH
is configured and that your $PATH includes $GOPATH/bin. **Note:** restart your
terminal session or reload the config file (e.g. `. ~/.zprofile`) after making
this change.

```bash
# in your .zprofile, .bash_profile, .zshrc or similar shell config
export GOPATH="${HOME}/go"
export PATH="${PATH}:${GOPATH}/bin"
```

Run the bootstrap script:

```bash
python3 bootstrap.py
```

Answer N (default) when asked to clone dependency repos over ssh. The bootstrap
script will `brew install` any necessary dependencies (Go, SDL2, etc.) and clone
support repos for the game (doodads, levelpacks, assets).

# WebAssembly

There is some **experimental** support for a WebAssembly build of Sketchy Maze
since the very early days. Early on, the game "basically worked" but performance
could be awful: playing levels was OK but clicking and dragging in the editor
would cause your browser to freeze. Then for a time, the game wouldn't even get
that far. Recently (December 2023), WASM performance seems much better but there
are strange graphical glitches:

* On the title screen, the example levels in the background load OK and their
  doodads will wander around and performance seems OK.
* But during Play Mode, only the menu bar draws but nothing else on the screen.
* In the Level Editor, the entire screen is white BUT tooltips will appear and
  the menu bar can be clicked on (blindly) and the drop-down menus do appear.
  Some popups like the Palette Editor can be invoked and draw to varying degrees
  of success.

Some tips to get a WASM build to work:

* For fonts: symlink it so that ./wasm/fonts points to ./assets/fonts.
* You may need an updated wasm_exec.js shim from Go. On Fedora,
  `dnf install golang-misc` and `cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .`
  from the wasm/ folder.
* Run `make wasm` to build the WASM binary and `make wasm-serve` to run a simple
  Go web server to serve it from.

# Build Tags

Go build tags used by this game:

## doodad

This tag is used when building the `doodad` command-line tool.

It ensures that the embedded bindata assets (built-in doodads, etc.) do not
need to be bundled into the doodad binary, but only the main game binary.

## dpp

The dpp tag stands for Doodle++ and is used for official commercial builds of
the game. Doodle++ builds include additional code not found in the free & open
source release of the game engine.

This build tag should be set automatically by the Makefile **if** the deps/
folder has a git clone of the dpp project. The bootstrap.py script will clone
the dpp repo **if** you use SSH to clone dependencies: so you will need SSH
credentials to the upstream git server. It basically means that third-party
users who download the open source release will not have the dpp dependency,
and will not build dpp copies of the game.

If you _do_ have the dpp dependency, you can force build (and run) FOSS
versions of the game via the Makefile commands `make build-free`,
`make run-free` or `make dist-free` which are counterparts to the main make
commands but which deliberately do not set the dpp build tag.

In source code, files ending with `_dpp.go` and `_foss.go` are conditionally
compiled depending on this build tag.

How to tell whether your build of Sketchy Maze is Doodle++ include:

* The version string on the title screen.
    * FOSS builds (not dpp) will say "open source" in the version.
    * DPP builds may say "shareware" if unregistered or just the version.
