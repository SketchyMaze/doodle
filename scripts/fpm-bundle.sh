#!/bin/bash

# fpm-bundle: create bundles for the app.

# Add the user-level "gem install fpm" to the $PATH.
# Might need fixing over time.
export PATH="$PATH:$HOME/.local/share/gem/ruby/3.0.0/bin"

INSTALL_ROOT="/opt/sketchy-maze"
LAUNCHER_FILENAME="etc/linux/net.kirsle.ProjectDoodle.desktop"
LAUNCHER_ROOT="/usr/share/applications" # Where the .desktop file goes.
ICON_ROOT="/usr/share/icons/hicolor/"

# Find out how many levels up we need to go, so this
# script can run from either of these locations:
# ./dist/sketchymaze-$version/
# ./dist/stage/$version/linux/
UPLEVELS="."
if [[ -f "../../${LAUNCHER_FILENAME}" ]]; then
  # run from a ./dist/x folder.
  UPLEVELS="../.."
elif [[ -f "../../../../${LAUNCHER_FILENAME}" ]]; then
  # run from a release stage folder
  UPLEVELS="../../../.."
else
  echo Did not find ${LAUNCHER_FILENAME} relative to your working directory.
  echo Good places to run this script include:
  echo " * ./dist/sketchymaze-\$version/  (as in 'make dist')"
  echo " * ./dist/stage/\$version/linux/  (as in 'make release')"
  exit 1
fi

VERSION=`egrep -e 'Version\s+=' ${UPLEVELS}/pkg/branding/branding.go | head -n 1 | cut -d '"' -f 2`
LAUNCHER_FILE="${UPLEVELS}/${LAUNCHER_FILENAME}"

if [[ ! -f "./sketchymaze" ]]; then
	echo Run this script from the directory containing the Doodle binary.
	echo This is usually at /dist/doodle-VERSION/ relative to the git root.
	exit 1
fi

if [[ ! -f "$LAUNCHER_FILE" ]]; then
	echo "Didn't find Linux desktop launcher relative to current folder."
	echo "I looked at $LAUNCHER_FILE."
	exit 1
fi

# Clean previous artifacts.
rm *.rpm *.deb

# Create the root structure.
mkdir -p root
mkdir -p root$INSTALL_ROOT root$LAUNCHER_ROOT
cp * root$INSTALL_ROOT/
cp $LAUNCHER_FILE root$LAUNCHER_ROOT/

# Copy icons in.
mkdir -p root$ICON_ROOT/{256x256,128x128,64x64,32x32,16x16}/apps
cp ${UPLEVELS}/etc/icons/256.png "root${ICON_ROOT}256x256/apps/project-doodle.png"
cp ${UPLEVELS}/etc/icons/128.png "root${ICON_ROOT}128x128/apps/project-doodle.png"
cp ${UPLEVELS}/etc/icons/64.png "root$ICON_ROOT/64x64/apps/project-doodle.png"
cp ${UPLEVELS}/etc/icons/32.png "root$ICON_ROOT/32x32/apps/project-doodle.png"
cp ${UPLEVELS}/etc/icons/16.png "root$ICON_ROOT/16x16/apps/project-doodle.png"

# Copy runtime package and guidebook
cp -r guidebook rtp "root$INSTALL_ROOT/"

echo =====================
echo Starting fpm package build.
echo =====================

# Handle all architectures! Default x86_64
RPM_ARCH="x86_64"
DEB_ARCH="x86_64"
case "$archs" in
  i?86)
    RPM_ARCH="i386"
    DEB_ARCH="i386"
    ;;
  aarch64)
    RPM_ARCH="aarch64"
    DEB_ARCH="arm64"
    ;;
esac

# RPM Package
fpm -C ./root -s dir -t rpm \
  -d SDL2 -d SDL2_ttf -a $RPM_ARCH \
  -n sketchy-maze -v ${VERSION} \
  --license="Copyright" \
  --maintainer=noah@kirsle.net \
  --description="Sketchy Maze - A drawing-based maze game." \
  --url="https://www.sketchymaze.com"

# Debian Package
fpm -C ./root -s dir -t deb \
  -d libsdl2-2.0 -d libsdl2-ttf-2.0 -d libsdl2-mixer-2.0 \
  -a $DEB_ARCH \
  -n sketchy-maze -v ${VERSION} \
  --license="Copyright" \
  --maintainer=noah@kirsle.net \
  --description="Sketchy Maze - A drawing-based maze game." \
  --url="https://www.sketchymaze.com"
