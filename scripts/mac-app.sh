#!/bin/bash

# mac-app: create Mac OS .app distribution. Run this script from the directory
# containing the doodad binaries (subdirectory of /dist from git repo root)

VERSION=`grep -e 'Version =' ../../pkg/branding/branding.go | head -n 1 | cut -d '"' -f 2`
INSTALL_ROOT="/opt/sketchy-maze"
APP_NAME="Project Doodle.app"
APP_FOLDER="../../etc/macos/$APP_NAME"
APP_CONTENTS="$APP_NAME/Contents"

if [[ ! -f "./doodle" ]]; then
	echo Run this script from the directory containing the Doodle binary.
	echo This is usually at /dist/doodle-VERSION/ relative to the git root.
	exit 1
fi

if [[ ! -d "$APP_FOLDER" ]]; then
	echo "Didn't find Mac .app template relative to current folder."
	echo "I looked at $APP_FOLDER."
	exit 1
fi

# Copy the Mac app template in to current folder.
echo Copying template app: $APP_FOLDER
cp -r "$APP_FOLDER" ./
mkdir -p "$APP_CONTENTS/MacOS"
mkdir -p "$APP_CONTENTS/Resources"

# Copy binaries to /MacOS
cp doodle doodad "$APP_CONTENTS/MacOS/"
cp *.* "$APP_CONTENTS/Resources/"
