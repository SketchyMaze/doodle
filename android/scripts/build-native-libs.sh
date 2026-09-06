#!/usr/bin/env bash
# Builds SDL2, SDL2_ttf and SDL2_mixer for each target Android ABI (via CMake
# + the NDK's toolchain file), then cross-compiles cmd/doodle-android against
# them as a cgo c-shared library. Run android/scripts/fetch-sdl-sources.sh
# first. See android/README.md for the full explanation of what's going on
# here and how to install the prerequisites (NDK, CMake).
set -euo pipefail

ANDROID_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DOODLE_DIR="$(cd "$ANDROID_DIR/.." && pwd)"
THIRD_PARTY_DIR="$ANDROID_DIR/third_party"
JNI_LIBS_DIR="$ANDROID_DIR/app/src/main/jniLibs"

: "${ANDROID_NDK_HOME:?Set ANDROID_NDK_HOME to your NDK install, e.g. \$ANDROID_HOME/ndk/30.0.16138531}"
: "${ANDROID_API:=21}" # Must be >= app/build.gradle's minSdkVersion.

TOOLCHAIN_FILE="$ANDROID_NDK_HOME/build/cmake/android.toolchain.cmake"
if [ ! -f "$TOOLCHAIN_FILE" ]; then
    echo "error: couldn't find $TOOLCHAIN_FILE -- is ANDROID_NDK_HOME correct?" >&2
    exit 1
fi

case "$(uname -s)" in
    Linux)  HOST_TAG=linux-x86_64 ;;
    Darwin) HOST_TAG=darwin-x86_64 ;;
    *) echo "error: unsupported host OS for this script: $(uname -s)" >&2; exit 1 ;;
esac
TOOLCHAIN_BIN="$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/$HOST_TAG/bin"

# Android ABI -> Go GOARCH, and (for armeabi-v7a) GOARM.
ABIS=(arm64-v8a armeabi-v7a x86_64)
declare -A GOARCH_FOR=( [arm64-v8a]=arm64 [armeabi-v7a]=arm [x86_64]=amd64 )
declare -A GOARM_FOR=( [armeabi-v7a]=7 )
# NDK clang binary prefix per ABI (see $NDK/toolchains/llvm/prebuilt/*/bin).
declare -A CLANG_TRIPLE_FOR=(
    [arm64-v8a]=aarch64-linux-android
    [armeabi-v7a]=armv7a-linux-androideabi
    [x86_64]=x86_64-linux-android
)
# The sysroot's per-ABI lib dir names (where libc++_shared.so lives) use the
# "plain" triple, which for 32-bit ARM differs from the clang driver name above.
declare -A SYSROOT_TRIPLE_FOR=(
    [arm64-v8a]=aarch64-linux-android
    [armeabi-v7a]=arm-linux-androideabi
    [x86_64]=x86_64-linux-android
)

build_sdl_lib() {
    local repo="$1" abi="$2" install_prefix="$3"; shift 3
    local build_dir="$THIRD_PARTY_DIR/$repo/build-$abi"
    echo "== building $repo for $abi =="
    cmake -S "$THIRD_PARTY_DIR/$repo" -B "$build_dir" -GNinja \
        -DCMAKE_TOOLCHAIN_FILE="$TOOLCHAIN_FILE" \
        -DANDROID_ABI="$abi" \
        -DANDROID_PLATFORM="android-$ANDROID_API" \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_INSTALL_PREFIX="$install_prefix" \
        -DCMAKE_PREFIX_PATH="$install_prefix" \
        -DCMAKE_FIND_ROOT_PATH="$install_prefix" \
        -DCMAKE_FIND_ROOT_PATH_MODE_PACKAGE=ONLY \
        -DCMAKE_FIND_ROOT_PATH_MODE_LIBRARY=BOTH \
        -DCMAKE_FIND_ROOT_PATH_MODE_INCLUDE=BOTH \
        -DBUILD_SHARED_LIBS=ON \
        -DANDROID_STL=c++_shared \
        -DCMAKE_POLICY_VERSION_MINIMUM=3.5 \
        "$@"
    cmake --build "$build_dir" --parallel
    cmake --install "$build_dir"
}

for abi in "${ABIS[@]}"; do
    install_prefix="$THIRD_PARTY_DIR/install/$abi"
    mkdir -p "$install_prefix"

    build_sdl_lib SDL "$abi" "$install_prefix"

    # NOTE: SDL_ttf and SDL_mixer's CMake options are prefixed SDL2TTF_/
    # SDL2MIXER_ (not SDLTTF_/SDLMIXER_ -- easy typo, and CMake silently
    # ignores unknown -D cache vars instead of erroring, so a wrong prefix
    # here fails silently by just falling back to upstream's defaults).
    build_sdl_lib SDL_ttf "$abi" "$install_prefix" \
        -DSDL2TTF_VENDORED=ON \
        -DSDL2TTF_SAMPLES=OFF

    # NOTE: unlike the others, there's no plain SDL2MIXER_OGG option -- OGG
    # Vorbis support is controlled by the SDL2MIXER_VORBIS string cache var
    # (STB/TREMOR/VORBISFILE), which already defaults to "STB" (a vendored,
    # dependency-free single-header decoder), so it's passed here only to
    # make that explicit rather than relying on an undocumented default.
    build_sdl_lib SDL_mixer "$abi" "$install_prefix" \
        -DSDL2MIXER_SAMPLES=OFF \
        -DSDL2MIXER_VENDORED=ON \
        -DSDL2MIXER_MP3=ON \
        -DSDL2MIXER_VORBIS=STB \
        -DSDL2MIXER_FLAC=OFF \
        -DSDL2MIXER_MOD=OFF \
        -DSDL2MIXER_MIDI=OFF \
        -DSDL2MIXER_OPUS=OFF \
        -DSDL2MIXER_WAVPACK=OFF

    echo "== copying $abi runtime .so files into the Gradle jniLibs dir =="
    mkdir -p "$JNI_LIBS_DIR/$abi"
    cp -v "$install_prefix"/lib/libSDL2.so \
          "$install_prefix"/lib/libSDL2_ttf.so \
          "$install_prefix"/lib/libSDL2_mixer.so \
          "$JNI_LIBS_DIR/$abi/"

    # SDL2_ttf (freetype/harfbuzz) and SDL2_mixer (libmpg123/libvorbis) are C++
    # under the hood and were built against the shared C++ runtime above
    # (-DANDROID_STL=c++_shared), so it has to ship in the APK too, once per
    # ABI, or you get "library libc++_shared.so not found" at load time.
    sysroot_triple="${SYSROOT_TRIPLE_FOR[$abi]}"
    cp -v "$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/$HOST_TAG/sysroot/usr/lib/$sysroot_triple/libc++_shared.so" \
          "$JNI_LIBS_DIR/$abi/"

    echo "== cross-compiling cmd/doodle-android for $abi =="
    triple="${CLANG_TRIPLE_FOR[$abi]}"
    cc="$TOOLCHAIN_BIN/${triple}${ANDROID_API}-clang"
    if [ ! -x "$cc" ]; then
        echo "error: expected NDK clang at $cc (got the wrong ANDROID_API or NDK version?)" >&2
        exit 1
    fi

    (
        cd "$DOODLE_DIR"
        env \
            CGO_ENABLED=1 \
            GOOS=android \
            GOARCH="${GOARCH_FOR[$abi]}" \
            GOARM="${GOARM_FOR[$abi]:-}" \
            CC="$cc" \
            CGO_CFLAGS="-I$install_prefix/include -I$install_prefix/include/SDL2" \
            CGO_LDFLAGS="-L$install_prefix/lib -lSDL2 -lSDL2_ttf -lSDL2_mixer" \
            go build -buildmode=c-shared -trimpath \
                -o "$JNI_LIBS_DIR/$abi/libdoodle.so" \
                ./cmd/doodle-android
    )
    # go build -buildmode=c-shared also emits a libdoodle.h we don't need.
    rm -f "$JNI_LIBS_DIR/$abi/libdoodle.h"

    echo "== $abi done =="
done

echo
echo "All ABIs built. jniLibs now contains:"
find "$JNI_LIBS_DIR" -name '*.so' | sort
echo
echo "Next: cd android && ./gradlew assembleDebug"
