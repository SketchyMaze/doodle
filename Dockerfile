##
# Fully build and distribute Linux and Windows binaries for Project: Doodle.
#
# This is designed to be run from a fully initialized Doodle environment
# (you had run the bootstrap.py for your system, and the doodads and
# levelpacks are installed in the assets/ folder, and `make dist` would
# build a release quality game for your local machine).
#
# It will take your working directory (minus any platform-specific artifacts
# and git repos cloned in to your deps/ folder) and build them from a sane
# Debian base and generate full release artifacts for:
#
# - Linux (x86_64 and i686) as .rpm, .deb, .flatpak and .tar.gz
# - Windows (64-bit and 32-bit) as .zip
#
# Artifact outputs will be in the dist/mw/ folder.
##

FROM debian:latest AS build64
ENV GOPATH /go
ENV GOPROXY direct
ENV PATH /opt/go/bin:/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/go/bin

# Install all dependencies.
RUN apt update && apt -y install git zip tar libsdl2-dev libsdl2-ttf-dev \
    libsdl2-mixer-dev gcc-mingw-w64-x86-64 gcc make wget \
    flatpak-builder ruby-dev gcc rpm libffi-dev \
    ruby-dev ruby-rubygems rpm libffi-dev rsync file
RUN gem install fpm; exit 0

# Download and install modern Go.
WORKDIR /root
RUN wget https://go.dev/dl/go1.21.4.linux-amd64.tar.gz -O go.tgz && \
    tar -xzf go.tgz && \
    cp -r go /opt/go

# Add some cacheable directories to speed up Dockerfile trial-and-error.
ADD deps/vendor /SketchyMaze/deps/vendor

# MinGW setup for Windows executable cross-compile.
WORKDIR /SketchyMaze/deps/vendor/mingw-libs
RUN for i in *.tar.gz; do tar -xzvf $i; done
RUN cp -r SDL2-2.0.9/x86_64-w64-mingw32 /usr && \
    cp -r SDL2_mixer-2.0.4/x86_64-w64-mingw32 /usr && \
    cp -r SDL2_ttf-2.0.15/x86_64-w64-mingw32 /usr
RUN mkdir -p /usr/lib/golang/pkg/windows_amd64
WORKDIR /SketchyMaze
RUN mkdir -p bin && cp deps/vendor/DLL/*.dll bin/

# Add the current working directory (breaks the docker cache every time).
ADD . /SketchyMaze

# Fetch the guidebook.
# RUN sh -c '[[ ! -d ./guidebook ]] && wget -O - https://download.sketchymaze.com/guidebook.tar.gz | tar -xzvf -'

# Use go-winres on the Windows exe (embed application icons)
RUN go install github.com/tc-hib/go-winres@latest && go-winres make

# Install Go dependencies and do the thing:
# - builds the program for Linux
# - builds for Windows via MinGW
# - runs `make dist/` creating an uber build for both OS's
# - runs release.sh to carve out the Linux and Windows versions and
#   zip them all up nicely.
RUN make setup && make from-docker64

# Collect the build artifacts.
RUN mkdir -p artifacts && cp -rv dist/release ./artifacts/

###
# 32-bit Dockerfile version of the above
###
FROM i386/debian:latest AS build32

ENV GOPATH /go
ENV GOPROXY direct
ENV PATH /opt/go/bin:/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/go/bin

# Dependencies, note the w64-i686 difference to the above
RUN apt update && apt -y install git zip tar libsdl2-dev libsdl2-ttf-dev \
    libsdl2-mixer-dev gcc-mingw-w64-i686 gcc make wget \
    flatpak-builder ruby-dev gcc rpm libffi-dev \
    ruby-dev ruby-rubygems rpm libffi-dev rsync file
RUN gem install fpm; exit 0

# Download and install modern Go.
WORKDIR /root
RUN wget https://go.dev/dl/go1.19.3.linux-386.tar.gz -O go.tgz && \
    tar -xzf go.tgz && \
    cp -r go /opt/go

COPY --from=build64 /SketchyMaze /SketchyMaze

# MinGW setup for Windows executable cross-compile.
WORKDIR /SketchyMaze/deps/vendor/mingw-libs
RUN for i in *.tar.gz; do tar -xzvf $i; done
RUN cp -r SDL2-2.0.9/i686-w64-mingw32 /usr && \
    cp -r SDL2_mixer-2.0.4/i686-w64-mingw32 /usr && \
    cp -r SDL2_ttf-2.0.15/i686-w64-mingw32 /usr
RUN mkdir -p /usr/lib/golang/pkg/windows_386
WORKDIR /SketchyMaze
RUN mkdir -p bin && cp deps/vendor/DLL-32bit/*.dll bin/

# Do the thing.
RUN make setup && make from-docker32

# Collect the build artifacts.
RUN mkdir -p artifacts && cp -rv dist/release ./artifacts/

###
# Back to (64bit) base for the final CMD to copy artifacts out.
###
FROM debian:latest

COPY --from=build32 /SketchyMaze /SketchyMaze
CMD ["cp", "-r", "-v", \
    "/SketchyMaze/artifacts/release/", \
    "/mnt/export/"]