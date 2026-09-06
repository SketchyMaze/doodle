#!/usr/bin/env bash
# Fetches the SDL2 / SDL2_ttf / SDL2_mixer C sources (pinned to known-good
# tags) that build-native-libs.sh compiles for Android, and copies the stock
# SDL Java glue (org.libsdl.app.*) into the Gradle project.
#
# Android has no system package manager to `apt install libsdl2-dev` from
# (unlike the desktop Linux build), so we build SDL2 and friends from source
# ourselves, once per ABI, with the NDK toolchain. Re-run this whenever you
# bump the pinned versions below; it's safe to run repeatedly (rm -rf's its
# own checkout dirs first).
set -euo pipefail

# Keep these three in lockstep with each other (SDL2_ttf/SDL2_mixer track
# SDL2 minor releases) and with whatever the go-sdl2 version in doodle's
# go.mod expects at the C API level -- see android/README.md.
SDL_VERSION=release-2.30.9
SDL_TTF_VERSION=release-2.22.0
SDL_MIXER_VERSION=release-2.8.0

ANDROID_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
THIRD_PARTY_DIR="$ANDROID_DIR/third_party"
JAVA_GLUE_DEST="$ANDROID_DIR/app/src/main/java/org/libsdl/app"

mkdir -p "$THIRD_PARTY_DIR"
cd "$THIRD_PARTY_DIR"

clone() {
    local name="$1" tag="$2" url="$3"
    if [ -d "$name" ]; then
        echo "-- $name already present, skipping clone (delete $THIRD_PARTY_DIR/$name to re-fetch)"
        return
    fi
    echo "-- cloning $name @ $tag"
    git clone --depth 1 --branch "$tag" "$url" "$name"
}

clone SDL "$SDL_VERSION" https://github.com/libsdl-org/SDL.git
clone SDL_ttf "$SDL_TTF_VERSION" https://github.com/libsdl-org/SDL_ttf.git
clone SDL_mixer "$SDL_MIXER_VERSION" https://github.com/libsdl-org/SDL_mixer.git

# SDL_ttf and SDL_mixer vendor their own dependencies (freetype/harfbuzz,
# mpg123/libvorbis) as submodules under external/.
for repo in SDL_ttf SDL_mixer; do
    echo "-- fetching $repo submodules"
    (cd "$repo" && git submodule update --init --recursive --depth 1)
done

echo "-- copying SDL's Java glue (org.libsdl.app) into the Gradle project"
mkdir -p "$JAVA_GLUE_DEST"
cp -v SDL/android-project/app/src/main/java/org/libsdl/app/*.java "$JAVA_GLUE_DEST/"

echo
echo "Done. Native sources are in $THIRD_PARTY_DIR."
echo "Next: run android/scripts/build-native-libs.sh"
