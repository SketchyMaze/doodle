#!/bin/bash

# Script to build an AppImage.
# Run it like `ARCH=x86_64 make appimage`
# It will fetch your appimagetool-x86_64.AppImage program to build the appimage.

if [[ ! -d "./dist/sketchymaze-latest" ]]; then
    echo "error: run make dist before make appimage"
    exit 1
fi

if [[ "$ARCH" == "" ]]; then
    echo "You should set ARCH=x86_64 (or your platform for AppImage output)"
    exit 1
fi

APPDIR="./dist/AppDir"
LAUNCHER="./scripts/appimage-launcher.sh"
DESKTOP="./scripts/appimage.desktop"
ICON_IMG="./etc/icons/256.png"
ICON_VECTOR="./etc/icons/orange-128.svg"

APP_RUN="$APPDIR/AppRun"
DIR_ICON="$APPDIR/sketchymaze.svg"

APPIMAGETOOL="appimagetool-$ARCH.AppImage"
if [[ ! -f "./$APPIMAGETOOL" ]]; then
    echo "Downloading appimagetool"
    wget "https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-$ARCH.AppImage"
    chmod a+x $APPIMAGETOOL
fi

# Clean start
if [[ -d "$APPDIR" ]]; then
    echo "Note: Removing previous $APPDIR"
    rm -rf "$APPDIR"
fi

echo "Creating $APPDIR"
mkdir $APPDIR

# Entrypoint script (AppRun)
cp "./scripts/appimage-launcher.sh" "$APPDIR/AppRun"
chmod +x "$APPDIR/AppRun"

# .DirIcon PNG for thumbnailers
cp "./etc/icons/256.png" "$APPDIR/.DirIcon"
cp "./etc/icons/orange-128.svg" "$APPDIR/sketchymaze.svg"

# .desktop launcher
cp "./scripts/appimage.desktop" "$APPDIR/sketchymaze.desktop"

# Everything else
rsync -av "./dist/sketchymaze-latest/" "$APPDIR/"

echo "Making AppImage..."
cd $APPDIR
../../$APPIMAGETOOL $(pwd)