#!/bin/bash

# release.sh:
# Run this AFTER `make dist` and it will carve up the dist/
# folder into Linux and Windows versions and zip them up
# for release.
#
# For Mac OS releases, after `make dist` run `release-osx.sh`
# instead.

# Add the user-level "gem install fpm" to the $PATH.
# Might need fixing over time.
export PATH="$PATH:$HOME/.local/share/gem/ruby/3.0.0/bin"

VERSION=`egrep -e 'Version\s+=' ./pkg/branding/branding.go | head -n 1 | cut -d '"' -f 2`
DIST_PATH="$(pwd)/dist/sketchymaze-${VERSION}"
RELEASE_PATH="$(pwd)/dist/release/${VERSION}"
STAGE_PATH="$(pwd)/dist/stage/${VERSION}"

if [[ ! -d $DIST_PATH ]]; then
	echo Run this script from the root of the game repository, such that
    echo ./dist/sketchymaze-${VERSION} exists. Run 'make mingw' and/or
    echo 'make dist' to create this directory before running this script.
	exit 1
fi

# Clean previous artifacts.
if [[ -d $RELEASE_PATH ]]; then
    rm -rf $RELEASE_PATH
fi
if [[ -d $STAGE_PATH ]]; then
    rm -rf $STAGE_PATH
fi

# Release scripts by operating system.
linux() {
    # Check for Linux binaries in the dist folder.
    if [[ ! -f "${DIST_PATH}/doodad" ]]; then
        echo No Linux binaries found, skipping Linux release.
        return
    fi

    # Prepare the Linux release.
    mkdir -p ${STAGE_PATH} "${RELEASE_PATH}/linux"
    cp -r $DIST_PATH "${STAGE_PATH}/linux"
    cd "$STAGE_PATH/linux"

    # Remove Windows artifacts.
    rm *.exe *.dll

    # Tar it.
    tar -czvf "${RELEASE_PATH}/linux/sketchymaze-${VERSION}-linux-64bit.tar.gz" .

    # fpm it.
    ../../../../scripts/fpm-bundle.sh
    cp *.rpm *.deb "${RELEASE_PATH}/linux/"

    # return
    cd -
}
windows() {
    # Check for Windows binaries in the dist folder.
    if [[ ! -f "${DIST_PATH}/doodad.exe" ]]; then
        echo No Windows binaries found, skipping Windows release.
        return
    fi

    # Prepare the Windows release.
    mkdir -p ${STAGE_PATH} "${RELEASE_PATH}/windows"
    cp -r $DIST_PATH "${STAGE_PATH}/windows"
    cd "$STAGE_PATH/windows"

    # Remove Linux artifacts.
    rm sketchymaze doodad

    # Zip it.
    zip -r "${RELEASE_PATH}/windows/sketchymaze-${VERSION}-windows-64bit.zip" .
    cd -
}

linux
windows