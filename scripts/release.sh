#!/bin/bash

# release.sh:
# Run this AFTER `make dist` and it will carve up the dist/
# folder into Linux and Windows versions and zip them up
# for release.
#
# If run on a Mac OS system, it runs only the Mac OS code
# path; it does not expect any Windows .exe files or .dll's,
# the dist folder should be a Mac only build of the game.
#
# Add the user-level "gem install fpm" to the $PATH.
# Might need fixing over time.
export PATH="$PATH:$HOME/.local/share/gem/ruby/3.0.0/bin"

VERSION=`egrep -e 'Version\s+=' ./pkg/branding/branding.go | head -n 1 | cut -d '"' -f 2`
DIST_PATH="$(pwd)/dist/sketchymaze-${VERSION}"
RELEASE_PATH="$(pwd)/dist/release/${VERSION}"
STAGE_PATH="$(pwd)/dist/stage/${VERSION}"

# Handle all architectures! Default x86_64
# Used in zipfiles for Linux and Windows (fpm-bundle.sh has its own logic)
ARCH_LABEL="64bit"
case "$archs" in
  i?86) ARCH_LABEL="32bit" ;;
  aarch64) ARCH_LABEL="aarch64" ;;
esac

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

# Check that we will bundle the Guidebook with the release.
if [[ ! -d "./guidebook" ]]; then
    echo Guidebook not found locally! Should I download the latest or proceed without it?
    echo Type 'y' to download the guidebook or press enter to skip.
    read ANSWER
    if [[ "$ANSWER" == "y" ]]; then
        wget https://download.sketchymaze.com/guidebook.tar.gz && tar -xzvf guidebook.tar.gz
    fi
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
    tar -czvf "${RELEASE_PATH}/linux/sketchymaze-${VERSION}-linux-${ARCH_LABEL}.tar.gz" .

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
    zip -r "${RELEASE_PATH}/windows/sketchymaze-${VERSION}-windows-${ARCH_LABEL}.zip" .
    cd -
}
macos() {
    # Check for OSX binaries in the dist folder.
    if [[ ! -f "${DIST_PATH}/doodad" ]]; then
        echo No binaries found, skipping Mac release.
        return
    fi

    # Prepare the OSX release.
    mkdir -p ${STAGE_PATH} "${RELEASE_PATH}/macos"
    cp -r $DIST_PATH "${STAGE_PATH}/macos"
    cd "$STAGE_PATH/macos"

    # Zip it.
    zip -r "${RELEASE_PATH}/macos/sketchymaze-${VERSION}-macos-x64.zip" .

    # Create the .app bundle.
    ../../../../scripts/mac-app.sh

    # Remove redundant Mac binaries from stage folder.
    rm ./sketchymaze ./doodad
    hdiutil create "${RELEASE_PATH}/macos/sketchymaze-${VERSION}-macOS-x64.dmg" \
        -ov -volname "SketchyMaze" -fs HFS+ -srcfolder $(pwd)

    cd -
}

if [[ `uname` == "Darwin" ]]; then
    macos
else
    linux
    windows
fi
