#!/bin/bash

# Script to build an AppImage.
#
# Dependencies:
# * appimage-builder: a Python module, so pip install -r requirements.txt

if ! command -v appimage-builder &> /dev/null;then
    echo "appimage-builder not found; run pip install -r requirements.txt"
    exit 1
fi

if [[ ! -d "./dist/sketchymaze-latest" ]]; then
    echo "error: run make dist before make appimage"
    exit 1
fi

APPDIR="./dist/AppDir"
LAUNCHER="./scripts/appimage-launcher.sh"
DESKTOP="./scripts/appimage.desktop"
ICON_IMG="./etc/icons/256.png"
ICON_VECTOR="./etc/icons/orange-128.svg"

APP_RUN="$APPDIR/AppRun"
DIR_ICON="$APPDIR/sketchymaze.svg"

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
appimagetool $(pwd)