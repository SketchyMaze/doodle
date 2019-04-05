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

## Windows

TBD.
