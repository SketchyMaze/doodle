#!/bin/bash

# fpm-bundle: create bundles for the app.

VERSION=`egrep -e 'Version\s+=' ../../pkg/branding/branding.go | head -n 1 | cut -d '"' -f 2`
INSTALL_ROOT="/opt/project-doodle"
LAUNCHER_FILE="../../etc/linux/net.kirsle.ProjectDoodle.desktop"
LAUNCHER_ROOT="/usr/share/applications" # Where the .desktop file goes.
ICON_ROOT="/usr/share/icons/hicolor/"

if [[ ! -f "./doodle" ]]; then
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
cp ../../etc/icons/256.png "root${ICON_ROOT}256x256/apps/project-doodle.png"
cp ../../etc/icons/128.png "root${ICON_ROOT}128x128/apps/project-doodle.png"
cp ../../etc/icons/64.png "root$ICON_ROOT/64x64/apps/project-doodle.png"
cp ../../etc/icons/32.png "root$ICON_ROOT/32x32/apps/project-doodle.png"
cp ../../etc/icons/16.png "root$ICON_ROOT/16x16/apps/project-doodle.png"

echo =====================
echo Starting fpm package build.
echo =====================

# RPM Package
fpm -C ./root -s dir -t rpm \
  -d SDL2 -d SDL2_ttf -a x86_64 \
  -n project-doodle -v ${VERSION} \
  --license="Copyright" \
  --maintainer=noah@kirsle.net \
  --description="Project: Doodle - A drawing-based maze game." \
  --url="https://www.kirsle.net/doodle"

# Debian Package
fpm -C ./root -s dir -t deb \
  -d libsdl2 -d libsdl2-ttf -a x86_64 \
  -n project-doodle -v ${VERSION} \
  --license="Copyright" \
  --maintainer=noah@kirsle.net \
  --description="Project: Doodle - A drawing-based maze game." \
  --url="https://www.kirsle.net/doodle"
