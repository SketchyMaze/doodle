# Building Doodle

Makefile commands for Linux:

* `make setup`: install Go dependencies and set up the build environment
* `make build`: build the Doodle and Doodad binaries to the `bin/` folder.
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

## Linux

Dependencies are Go, SDL2 and SDL2_ttf:

```bash
# Fedora
sudo dnf -y install golang SDL2-devel SDL2_ttf-devel

# Ubuntu and Debian
sudo apt -y install golang libsdl2-dev libsdl2-ttf-dev
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
