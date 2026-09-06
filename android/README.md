# SketchyMaze for Android

This folder is a Gradle/Android Studio project that wraps the existing Go +
SDL2 game (this whole `doodle/` repo) into an Android APK. It does **not**
reimplement the game -- it loads the real game as a native shared library
(`libdoodle.so`) inside SDL2's stock Android activity.

- [Quick Start](#quick-start)
- [How it fits together](#how-it-fits-together)
- [Directory layout](#directory-layout)
- [Prerequisites](#prerequisites)
- [Build steps](#build-steps)
  1. [Fetch and build the native SDL2 libraries](#1-fetch-and-build-the-native-sdl2-libraries)
  2. [Build the APK](#2-build-the-apk)
  3. [Install and run](#3-install-and-run)
- [Assets and save data](#assets-and-save-data)
- [Screen resolution / DPI scaling](#screen-resolution--dpi-scaling)
- [Signing a release build](#signing-a-release-build)
- [Known gaps / TODO](#known-gaps--todo)
- [Troubleshooting](#troubleshooting)
- [Updating SDL2 versions](#updating-sdl2-versions)

## Quick Start

If you just want to get building without reading the rest of this file
first, `bootstrap.py` and the `Makefile` automate everything below.

```sh
cd android
make setup            # one-time: installs/checks host deps, fetches SDL2 sources
make debug-install     # full rebuild (native libs + APK) and install to a connected device
```

`bootstrap.py` installs what it can via your system package manager
(cmake, ninja, git, go), detects Android Studio/the SDK/NDK (asking where
they are if they're not at the usual default paths -- and remembering the
answer, so it won't ask again), fetches the SDL2/SDL2_ttf/SDL2_mixer
sources, and writes an `android/.env` file the `Makefile` and your own
shell (`source .env`) can both use. It's safe to re-run: `make setup`
again later just reports what's already in place. See `bootstrap.py`'s own
docstring, or the [Prerequisites](#prerequisites) section below, for what
it can't do for you (mainly: installing Android Studio itself, and any SDK
components its own GUI SDK Manager has to install).

| Command | What it does |
|---|---|
| `make setup` | Bootstrap host deps + fetch SDL2 sources (safe to re-run) |
| `make setup-refetch` | Also re-fetch SDL2/SDL2_ttf/SDL2_mixer even if already present (e.g. after bumping the pinned version in `scripts/fetch-sdl-sources.sh`) |
| `make doctor` | Non-interactive `make setup` that only reports what's missing |
| `make native` | Rebuild `libdoodle.so` + SDL2/SDL2_ttf/SDL2_mixer for every ABI (`scripts/build-native-libs.sh`) -- the slow step, only needed after Go changes |
| `make debug` | Full rebuild: `make native` + a debug APK |
| `make debug-install` | `make debug` + force-stop + install -- the one-command loop after changing **Go** code |
| `make apk` | Just the debug APK, skipping the native rebuild -- fast |
| `make apk-install` | `make apk` + force-stop + install -- the one-command loop after changing **Java/Gradle/resources** only |
| `make build` | Full rebuild in release mode (needs a signing config -- see [Signing a release build](#signing-a-release-build)) |
| `make install` | `make build` + force-stop + install |
| `make run` | Launch the already-installed app without rebuilding |
| `make stop` | Force-stop the app |
| `make uninstall` | Uninstall the app (and its save data) from the device |
| `make log` | `adb logcat -s SDL:V` -- see [Assets and save data](#assets-and-save-data) for where the game's *own* (non-SDL) logs end up instead |
| `make clean` | Remove Gradle's build outputs |
| `make distclean` | Also remove the fetched SDL2 sources, built `.so` files, and `.env` -- next `make setup` starts from scratch |

## How it fits together

SDL2 on Android works by shipping a stock Java `Activity`
(`org.libsdl.app.SDLActivity`, from the SDL2 source tree) that creates the
window/`SurfaceView`, pumps Android lifecycle and input events into SDL's
C event queue, and then `dlopen()`s **your** native library and calls a
named C entry point in it (conventionally `SDL_main`) on a background
thread.

For a normal SDL2 game that entry point is written in C/C++. Here it's Go:
this repo already uses [go-sdl2](https://github.com/veandco/go-sdl2), which
supports Android, and cgo lets a Go program export a C symbol named
`SDL_main` directly. So the pipeline is:

```
cmd/doodle-android/main_android.go   (Go, cgo, "//export SDL_main")
        |  GOOS=android GOARCH=arm64 CGO_ENABLED=1 go build -buildmode=c-shared
        v
android/app/src/main/jniLibs/arm64-v8a/libdoodle.so
        |  packaged as-is by Gradle (no C/C++ compilation happens in Gradle)
        v
MainActivity.java (extends SDLActivity)
        |  System.loadLibrary("doodle"), then JNI-calls exported SDL_main()
        v
pkg/doodle.New(...).Run()   -- the same game loop as the desktop build
```

The only Android-specific Go code is `cmd/doodle-android/main_android.go`
(outside this folder, alongside `cmd/doodle`, `cmd/doodad`, etc., so it can
import the rest of the module normally). Everything else -- rendering,
levels, doodads, scripting, the level editor -- is the exact same code that
runs on desktop. `pkg/native/touch_screen.go` already detects touch input at
runtime, so the UI already adapts somewhat to a touchscreen-only device.

Gradle itself never compiles any C/C++/Go code. `app/build.gradle` just
packages whatever `.so` files it finds under
`app/src/main/jniLibs/<abi>/`; a separate build script
(`scripts/build-native-libs.sh`) is what actually produces those files.

## Directory layout

```
android/
  README.md                    -- this file
  Makefile                     -- see "Quick Start" above
  bootstrap.py                 -- see "Quick Start" above
  .env                          -- NOT checked in; written by bootstrap.py
  settings.gradle, build.gradle, gradle.properties   -- top-level Gradle project
  local.properties.example     -- copy to local.properties, point at your SDK
  play-store-icon.png          -- 512x512 listing icon, generated from assets/icons/1024.png
  app/
    build.gradle               -- the "app" module (applicationId, SDK versions, etc.)
    proguard-rules.pro
    src/main/
      AndroidManifest.xml
      java/
        com/sketchymaze/doodle/MainActivity.java   -- extends SDLActivity
        org/libsdl/app/...                          -- NOT checked in, see below
      res/
        values/strings.xml, styles.xml
        mipmap-*/ic_launcher*.png                    -- generated from assets/icons/
      jniLibs/<abi>/*.so                             -- NOT checked in, build output
  scripts/
    fetch-sdl-sources.sh        -- clones SDL2/SDL2_ttf/SDL2_mixer + copies their Java glue
    build-native-libs.sh        -- builds those C libs, then cross-compiles the Go code
  third_party/                  -- NOT checked in; created by fetch-sdl-sources.sh

cmd/doodle-android/
  main_android.go               -- the Go/cgo entry point (see "How it fits together")
```

Everything marked "NOT checked in" is either a large third-party source
checkout or a compiled binary; `.gitignore` excludes them and the scripts
regenerate them. This mirrors the fact that the desktop Linux build also
doesn't vendor `libSDL2.so` -- it links against whatever's installed on the
system. Android has no system package manager to install SDL2 from, so we
build our own copy instead, but the same principle (don't commit someone
else's compiled/generated output) applies.

## Prerequisites

`make setup` (see [Quick Start](#quick-start)) checks/installs most of this
automatically and tells you exactly what's missing for anything it can't
do itself. The rest of this section is the manual version, for reference
or if you'd rather not use it.

- **Android Studio** (easiest way to get the SDK, an emulator, and a working
  `local.properties`), or just the command-line `Android SDK` + `cmdline-tools`.
- **Android NDK**, r27 or newer -- not just "any recent NDK": r27 is where
  the linker started defaulting to 16 KB page-aligned native libraries,
  which newer Android versions/devices require (see the
  [Troubleshooting](#troubleshooting) entry on this if you're curious why
  that's a hard minimum and not just a suggestion). Install it via Android
  Studio's SDK Manager (SDK Tools tab -> NDK (Side by side)), or set
  `ANDROID_NDK_HOME` at whatever path you already have one -- this project
  is pinned to (and tested against) 30.0.16138531, set as `ndkVersion` in
  `app/build.gradle`.
- **CMake** and **Ninja** (Android Studio's SDK Manager can install CMake
  too; Ninja can come from your distro package manager, e.g.
  `sudo pacman -S ninja` / `apt install ninja-build`).
- **Go**, the same version this repo already requires (see `../go.mod`).
- **git**, to fetch the SDL2 sources.

You do *not* need Android Studio to do the actual build -- everything below
works from the command line -- but it's the easiest way to run/debug on an
emulator and to let it generate the Gradle wrapper (see below).

### Gradle wrapper

This folder intentionally does **not** commit `gradlew` / `gradlew.bat` /
`gradle/wrapper/gradle-wrapper.jar`, since the wrapper jar is a binary blob
that has to come from a real Gradle install. Generate it once:

```sh
cd android
gradle wrapper --gradle-version 8.7   # any Gradle you have installed can do this
```

or simply open the `android/` folder in Android Studio, which offers to do
the same thing automatically the first time you open the project.

## Build steps

`make debug-install` (see [Quick Start](#quick-start)) does everything in
this section in one command. What follows is what it's actually running,
if you want to run the steps individually or understand what's happening.

### 1. Fetch and build the native SDL2 libraries

The game currently uses three SDL2 libraries (see `deps/render/sdl` and
`deps/audio`): `SDL2`, `SDL2_ttf` and `SDL2_mixer` (with MP3 + OGG support,
for `.wav`/`.ogg`/`.mp3` sound effects and music -- see
`pkg/sound/sound_sdl.go`). `SDL2_image` is **not** needed; PNG loading goes
through Go's own `image/png`.

```sh
cd android
./scripts/fetch-sdl-sources.sh
```

This clones pinned tags of SDL2/SDL2_ttf/SDL2_mixer (and their vendored
dependencies, freetype/harfbuzz/mpg123/libvorbis) into `third_party/`, and
copies the stock SDL Java glue (`org.libsdl.app.*` -- `SDLActivity.java` and
friends) into `app/src/main/java/org/libsdl/app/`. That Java code is
intentionally *not* hand-copied into this repo by hand: it has to match the
native SDL2 version being built below, so it's fetched fresh alongside it
rather than risk the two drifting apart.

Then build the native libraries for each ABI and cross-compile the Go code
against them:

```sh
export ANDROID_NDK_HOME=/path/to/your/Sdk/ndk/30.0.16138531
./scripts/build-native-libs.sh
```

This builds `libSDL2.so`, `libSDL2_ttf.so`, `libSDL2_mixer.so` and
`libc++_shared.so` (the NDK's shared C++ runtime, needed because
SDL2_ttf/SDL2_mixer pull in C++ dependencies) for `arm64-v8a`,
`armeabi-v7a`, and `x86_64`, then runs:

```sh
CGO_ENABLED=1 GOOS=android GOARCH=<...> CC=<ndk-clang> \
    go build -buildmode=c-shared -o app/src/main/jniLibs/<abi>/libdoodle.so \
    ./cmd/doodle-android
```

once per ABI, with `CGO_CFLAGS`/`CGO_LDFLAGS` pointed at the SDL2 build
above. Everything ends up in `app/src/main/jniLibs/<abi>/`. Re-run this
script whenever you change any Go code -- it's much faster on repeat runs
since the SDL2/SDL2_ttf/SDL2_mixer CMake builds are incremental.

`arm64-v8a` is what real devices need (Google Play requires 64-bit native
code); `armeabi-v7a` covers older 32-bit devices; `x86_64` is only useful
for the Android emulator on an x86_64 host and can be dropped from release
builds if you want a smaller APK.

### 2. Build the APK

```sh
cd android
./gradlew assembleDebug
```

The APK comes out at `app/build/outputs/apk/debug/app-debug.apk`. Or just
open `android/` in Android Studio and hit Run -- it'll pick up the
`jniLibs` that `build-native-libs.sh` already produced.

### 3. Install and run

```sh
adb install -r app/build/outputs/apk/debug/app-debug.apk
```

or `./gradlew installDebug` with a device/emulator connected.

## Assets and save data

- **Game assets** (fonts, doodads, level packs, wallpapers, etc.) are
  already `//go:embed`-ded into the binary at compile time (see
  `assets/assets_embed.go`), so they need no special handling for Android --
  they're just part of `libdoodle.so` like they're part of the desktop
  binary.
- **Sound effects and music are the one exception**: `pkg/sound`
  (`SoundRoot`/`MusicRoot`) resolves them as plain relative filesystem
  paths into `rtp/` -- a runtime-package folder that, unlike `assets/`,
  ships as loose files next to the executable on desktop rather than being
  embedded, and is resolved against the process's working directory.
  Android has neither "next to the executable" nor a meaningful working
  directory, so those loads silently failed until this was accounted for.
  `rtp/embed_android.go` (an Android-only, build-tagged file, so it has no
  effect on the desktop/WASM builds) embeds `rtp/sfx` for this platform
  only; `cmd/doodle-android/main_android.go`'s `extractRTP()` unpacks it to
  the app's cache directory once at startup and repoints
  `sound.SoundRoot`/`MusicRoot` at the extracted copy -- reassigning those
  package vars directly rather than `os.Chdir()`ing the whole process,
  since nothing else in the codebase reads from a relative `"rtp"` path.
  If a `rtp/music` folder is ever added, update the `//go:embed` line in
  `rtp/embed_android.go` to include it too.
- **Save data / settings / user-created levels** go through
  `pkg/userdir`, which uses `github.com/kirsle/configdir` to find a
  per-OS config directory. `configdir` doesn't know about `GOOS=android`
  specifically, so it falls back to its generic XDG behavior, which reads
  `$HOME`/`$XDG_CONFIG_HOME`/`$XDG_CACHE_HOME` -- none of which exist by
  default in an Android app's process environment, and `pkg/userdir`'s
  `init()` (which resolves and creates all of its directories from those)
  runs at `dlopen()` time, before any Go code of ours gets a chance to run.
  `MainActivity.loadLibraries()` does set those three env vars (pointed at
  `getFilesDir()`) *before* `System.loadLibrary("doodle")` triggers that
  dlopen, which in principle should be timed correctly -- but in practice,
  even with Java's `Os.setenv()` confirmed (by logging) to succeed without
  throwing, the resulting directories never actually appeared on disk. The
  most likely explanation is that Go's runtime reads its own snapshot of
  the process environment once at startup rather than calling libc's
  `getenv()` live, so it never observed the change regardless of the
  ordering.

  Rather than keep chasing that, `cmd/doodle-android/main_android.go`'s
  `setupUserdir()` sidesteps environment variables entirely: it asks
  Android directly, via `SDL_AndroidGetInternalStoragePath()` (wrapped as
  `sdl2.AndroidGetInternalStoragePath()` in go-sdl2 -- the same JNI call
  that underlies `Context.getFilesDir()` on the Java side, just reached
  directly instead of round-tripped through an environment variable), and
  overwrites `pkg/userdir`'s exported path vars and re-creates its
  directories itself, as the very first thing `SDL_main()` does. This
  works because that JNI accessor only depends on the JNI environment
  SDLActivity sets up at library-load time (`nativeSetupJNI()`, which logs
  well before `SDL_main()` ever runs) -- not on `SDL_Init()` having run,
  and not on anything Java-side env-var plumbing.

## Screen resolution / DPI scaling

The game's UI (fonts, buttons, the level editor toolbox, ...) is laid out
in fixed pixel sizes tuned for an ordinary desktop monitor at ~96 DPI.
SDL's Android backend always reports the screen's *physical* pixel
resolution as the window size -- `SDL_CreateWindow`'s requested
width/height are silently discarded on Android (see
`third_party/SDL/src/video/android/SDL_androidwindow.c`: `window->w`/`h`
get overwritten from whatever the Java side reports, regardless of what's
passed to `sdl.New(...)` in `cmd/doodle-android/main_android.go`). A
modern phone panel (e.g. a Pixel 7 at 1080x2400) is commonly 2.5-3x denser
than a desktop monitor, so the same fixed-pixel UI shrinks to a third of
its intended size -- the mobile version of running an unscaled desktop
Linux UI on a 4K/Retina panel.

This is fixed entirely at the Android/Java layer, with no changes to any
shared Go code: `MainActivity.createSDLSurface()` returns
`ScaledSDLSurface` (`app/src/main/java/com/sketchymaze/doodle/
ScaledSDLSurface.java`) instead of the stock `SDLSurface`. It requests a
render buffer sized in device-independent pixels (physical pixels divided
by `DisplayMetrics.density`, the same scale factor Android uses
everywhere else) via `SurfaceHolder.setFixedSize()`, and the compositor
scales that smaller buffer back up to fill the real screen -- so SDL (and
the game) only ever sees a window sized like a normal desktop one, while
the physical display still renders at full resolution. It also corrects
touch-coordinate normalization, which stock `SDLSurface` derives from the
(now intentionally shrunk) buffer size rather than the real on-screen
touch position -- see the comments in `ScaledSDLSurface.java` for the
details of both.

This intentionally lives in our own package (`com.sketchymaze.doodle`)
rather than patching `org.libsdl.app.SDLSurface` directly, since that file
is re-copied from upstream SDL2 by `fetch-sdl-sources.sh` and would
clobber any in-place edit there on the next fetch.

If the UI still looks too small/large on your device, or you'd rather
target a fixed logical resolution instead of a density-derived one, that's
the file to tune.

**Getting the target size right requires computing it from the right
source.** `ScaledSDLSurface.surfaceChanged()` derives the target
width/height from `getWidth()`/`getHeight()` -- this `View`'s own measured
on-screen pixel size -- rather than from `mDisplay.getRealMetrics()` (the
`Display` object's metrics, which was the first version of this code).
That distinction matters because `mDisplay` reflects the *device's
physical rotation sensor*, while `getWidth()`/`getHeight()` reflects this
Activity's actual rendered orientation. Those normally agree, but they can
disagree if the app's orientation is locked (via
`android:screenOrientation`) to something other than whichever way the
phone happens to be held -- and when they disagree, the computed
width/height come out swapped relative to the real `SurfaceView` shape,
which is exactly what produces a squished, letterboxed-looking render
(fills most of one axis, a thin sliver of the other) rather than a clean
scaled-down copy of the desktop UI. This is also why `AndroidManifest.xml`
sets `android:screenOrientation="fullSensor"` rather than locking to one
orientation: the game already has its own portrait-friendly responsive
layout (see the `balance.Width`/`Height` overrides for phone-sized windows
in `cmd/doodle/main.go`), so there's no reason to fight it with an
orientation lock that also happens to reintroduce this exact class of bug.

**`getWidth()`/`getHeight()` are only a stable reference if the view is
actually laid out full-screen.** `SDLActivity.onCreate()` adds the surface
to its `RelativeLayout` with a bare `addView(mSurface)` -- no explicit
`LayoutParams` -- which makes `RelativeLayout` default it to
`WRAP_CONTENT` instead of filling the screen. Left alone, that breaks the
`getWidth()`/`getHeight()` fix above in a much worse way: instead of
reflecting the real screen size, they track something close to whatever
the *previous* `setFixedSize()` buffer was, so every `surfaceChanged()`
call (each device rotation, for instance) shrinks the buffer a bit more
than the last -- visible as the on-screen game getting smaller with every
rotation, eventually down to a sliver. `ScaledSDLSurface`'s constructor
works around this by giving itself explicit `MATCH_PARENT` `LayoutParams`
*before* `SDLActivity` ever adds it to the layout --
`ViewGroup.addView()` only falls back to `WRAP_CONTENT` defaults when the
child doesn't already have `LayoutParams` of its own, so this is enough to
fix it without touching `SDLActivity.java` itself.

**The very first frame after a cold launch needs a second safety net.**
On a live device rotation, by the time `surfaceChanged()` fires, the
window/insets/sensor-resolved orientation have already fully settled, so
`getWidth()`/`getHeight()` are reliable there. On a cold launch, the very
first `surfaceChanged()` call can race ahead of that settling -- e.g. it
can fire before immersive/fullscreen flags or the `fullSensor`-resolved
rotation have taken effect -- and see a transient, wrong-shaped size, with
nothing else naturally re-checking afterward once that first call resumes
native rendering. To catch that, `ScaledSDLSurface` also overrides
`onSizeChanged()`, a `View` callback that reliably fires whenever this
view's real measured size changes, independent of `surfaceChanged()`'s own
timing, and re-applies `setFixedSize()` there too if the target size has
changed. The actual size computation lives in one shared
`requestScaledBufferSize()` helper that both callbacks call into.

(As it turned out, the `onSizeChanged()` safety net above wasn't the real
fix for the cold-launch bug it was written for -- see the next entry for
what actually was.)

**SDL's own orientation-locking heuristic overrides the manifest, using
the *wrong* window size.** `SDL_CreateWindow()` calls into
`SDLActivity.setOrientationBis()` (via a native `setOrientation()` JNI
call) using the game's *original, un-scaled* requested window size --
`balance.Width`/`Height`, 1024x768 -- since that's what's passed to
`sdl.New(...)` before `ScaledSDLSurface` ever resizes anything. With a
resizable window and no `SDL_HINT_ORIENTATIONS` hint, the stock
implementation resolves 1024x768 (landscape-shaped) to
`SCREEN_ORIENTATION_FULL_USER` and calls `setRequestedOrientation()` with
it -- silently overriding `AndroidManifest.xml`'s
`android:screenOrientation="fullSensor"` at runtime, after our correctly
density-scaled portrait buffer had already been established.
`FULL_USER` behaves differently from `fullSensor`: if the user has
auto-rotate off, it locks to whatever orientation the system currently
considers "current" rather than the phone's live physical orientation --
which is what produced a wrong cold-launch orientation (visible as an
off-center, landscape-shaped title screen on a portrait launch) that only
self-corrected after the first live device rotation forced a fresh sensor
read.

`setOrientationBis()` is explicitly marked `/** This can be overridden */`
in SDL's own source, so `MainActivity` overrides it to just re-assert
`SCREEN_ORIENTATION_FULL_SENSOR` (matching the manifest) instead of
letting this heuristic pick something else based on a window size that
was never actually going to be the real on-screen size to begin with.

**`pkg/doodle.New()`'s width/height fields are only correct on desktop by
coincidence.** Even with the buffer size and orientation both right from
the very first frame, the title screen still came up landscape-shaped and
off-center on a cold portrait launch (identical to the exact same 411x838
buffer size looking correct after a rotation) -- so the remaining bug
wasn't in the Android layer at all. `doodle.New()` (`pkg/doodle.go`) seeds
its `width`/`height` fields from the hardcoded `balance.Width`/`Height`
(1024x768), and those fields only get corrected later by a genuine
`SDL_WINDOWEVENT_RESIZED` event inside `Doodle.Run()`'s own event loop.
On desktop this is a latent no-op: `SDL_CreateWindow()` actually creates a
window at exactly 1024x768, so the hardcoded default happens to already
be right. On Android, `SDL_CreateWindow()`'s requested size is always
overridden to match the real (density-scaled) window (see
`ScaledSDLSurface` above), so those fields are wrong from construction
until the first live resize event corrects them -- which is exactly what
produced a landscape-shaped initial layout on a portrait cold launch,
self-correcting only once an actual rotation fired a resize event.

Since there's already a public `Doodle.SetWindowSize(width, height int)`
method built for exactly this, `cmd/doodle-android/main_android.go` (not
any shared code) calls `game.SetWindowSize(engine.WindowSize())` right
after `game.SetupEngine()` -- which has already created the real window --
and before `game.Run()`, so the very first scene layout uses the correct
numbers instead of waiting on a resize event that, on a cold launch,
hasn't happened yet.

This one's a Go change, not Java, so it needs a native rebuild, not just a
Gradle one: `./scripts/build-native-libs.sh && ./gradlew assembleDebug`.

## Signing a release build

Android refuses to install *any* unsigned APK, even for local sideloading
-- so `app/build.gradle`'s `release` buildType is signed with
`signingConfigs.debug` by default (the Android Gradle Plugin's own
predefined config, pointing at the auto-generated
`~/.android/debug.keystore`). That's enough for `make build`/`make
install` to produce a working, installable `app-release.apk` for local
testing with zero setup -- but it's **not** appropriate for actually
distributing a release anywhere, since anyone with the (well-known, same
for every Android developer) debug keystore password could re-sign an APK
that Android would treat as an "update" to yours.

Before a real release, generate your own keystore:

```sh
keytool -genkey -v -keystore sketchymaze-release.keystore \
    -alias sketchymaze -keyalg RSA -keysize 2048 -validity 10000
```

then replace `signingConfig signingConfigs.debug` in `app/build.gradle`'s
`release` buildType with a `signingConfigs` block referencing that
keystore instead (see the [Android docs on signing]
(https://developer.android.com/studio/publish/app-signing)), and build
with `./gradlew bundleRelease` (for a Play Store `.aab`) or `make build`/
`./gradlew assembleRelease` (for a standalone signed `.apk`). Keep the
keystore itself out of git -- `.gitignore` already excludes
`android/*.keystore` and `android/*.jks`.

Also bump `versionCode`/`versionName` in `app/build.gradle` to match
`pkg/branding/branding.go`'s `Version` constant when you cut a release.

## Known gaps / TODO

This scaffold gets the game compiling and running as a real Android APK,
but a few things that made sense for a desktop app haven't been adapted
for a phone/tablet yet:

- **Native file dialogs -- reading works, saving doesn't yet.**
  `pkg/native/file_dialog_android.go`'s `OpenFile()` launches Android's
  Storage Access Framework picker (`ACTION_OPEN_DOCUMENT`, via JNI --
  `pkg/native/android/filepicker.go` on the Go side,
  `MainActivity.showFilePicker()`/`onActivityResult()` on the Java side)
  and copies the picked `content://` URI to a local cache file, since Go's
  `os.Open()` can't read those directly. `SaveFile()` returns an explicit
  "not supported" error instead of pretending to work: nothing in this
  codebase actually calls `native.SaveFile` today, and a real
  `ACTION_CREATE_DOCUMENT` implementation would need to copy the caller's
  bytes back out to the picked URI *after* they're written, which has no
  natural hook to attach to without a real caller to design it against.
- **"Open web browser" links** (`pkg/native/browser_android.go`) call
  `SDL_OpenURL()` directly via cgo (go-sdl2 doesn't wrap it, despite SDL's
  own `TODO.md` claiming otherwise) to fire an `Intent.ACTION_VIEW`. The
  in-game guidebook link (`OpenLocalURL(balance.GuidebookPath)`) redirects
  to the hosted copy (`branding.GuidebookURL`) instead, since the
  guidebook isn't bundled into the Android build the way `rtp/`'s audio
  is (see "Assets and save data" above) and a local `file://` URL
  couldn't usefully open in an external browser anyway. Other
  `OpenLocalURL` callers (the screenshots directory, an exported doodad's
  HTML docs) have no online equivalent to redirect to, so those just log
  a "not supported" warning instead.
- **On-screen keyboard.** `pkg/native.ShowKeyboard()`/`HideKeyboard()`
  (no-ops everywhere except Android, see `pkg/native/keyboard_android.go`)
  wrap `SDL_StartTextInput()`/`StopTextInput()`, which already trigger
  Android's soft keyboard on their own -- SDL's Android video backend
  wires that up internally, no JNI needed here at all. Wired into the dev
  shell (`pkg/shell.go`'s `Shell.open()`/`Close()`) and
  `Doodle.Prompt()`/`PromptPre()`. Two more things this needed once tested
  on a real device:
  - Typed characters weren't reaching the game at all. SDL's Android IME
    layer does synthesize a proper `SDL_KEYDOWN`+`SDL_KEYUP` pair per
    character (via `SDL_SendKeyboardUnicodeKey()`, called from
    `SDLInputConnection.commitText()` on the Java side) rather than only
    firing `SDL_TEXTINPUT` -- but with no real hold duration between the
    down and up, both landing in the same `Poll()` batch. `KeysDown()`
    (used by `Shell.Draw()`) only reflects keys *currently* held, so the
    key had already gone back up before the shell ever checked. Fixed at
    the `git.kirsle.net/go/render/event` level (`deps/render/` in this
    repo): a new `KeysPressed()`/`ResetKeysPressed()` pair tracks "had a
    keydown since the last `Poll()`" separately from `KeysDown()`'s
    continuous-hold state, reset once per `Poll()`
    (`deps/render/sdl/events.go`, `deps/render/canvas/events.go`) rather
    than accumulating forever. `Shell.Draw()` now reads from
    `KeysPressed()`. This is a real SDL2/Android behavior, not Android-
    specific to this game, so it's fixed at the shared render-engine
    level rather than worked around in `pkg/`.
  - The keyboard popped up **over** the game instead of the window
    resizing to make room, so the shell (drawn at the bottom of the
    screen) ended up hidden behind it.
    `android:windowSoftInputMode="adjustResize"` -- the normal fix for
    this -- doesn't reliably apply to a fullscreen/immersive activity like
    this one (see `SDLActivity.setWindowStyle()`'s `SYSTEM_UI_FLAG_*`
    flags), so `ScaledSDLSurface` now detects the keyboard manually (the
    classic `getWindowVisibleDisplayFrame()` height-comparison technique,
    portable back to this app's `minSdkVersion` unlike the newer
    `WindowInsets.Type.ime()` API) and resizes its own `LayoutParams`
    height to leave room for it -- which chains into the existing
    `onSizeChanged()`/`requestScaledBufferSize()` pipeline for free.
    Relatedly: `Doodle.Run()`'s main loop skips `Scene.Loop()` entirely
    while the shell is open (so typing doesn't also drive player
    movement/clicks in the background), which also meant a resize
    mid-shell never reached the current scene's own layout code (`Draw()`
    still ran every frame, but only re-rendering stale widget positions).
    Fixed by forwarding a sanitized `event.State` -- only `WindowResized`
    and the cursor position set, nothing interactive -- to `Scene.Loop()`
    specifically when `ev.WindowResized` is true while the shell is open.
    Since `PlayScene.Loop()` is now the only one of these that could
    actually advance gameplay simulation, and it happened to already
    return early on `WindowResized` before reaching that code, this
    couldn't have progressed gameplay in practice -- but relying on that
    incidentally was fragile, so `PlayScene.Paused()` (checked explicitly
    at the top of the simulation block) now makes it deliberate.
  - The Enter key didn't submit the prompt. Same root cause as the typed-
    character bug above, different field: `ev.Enter` is a raw "currently
    held" boolean (unrelated to the `KeysDown()`/`KeysPressed()` machinery)
    that a synthesized keydown+keyup pair in the same batch cleared before
    `keybind.Enter()` -- which already treated it as edge-triggered
    (consume-and-clear on read) -- ever saw it true. Fixed the same way:
    `Poll()` now resets `s.Enter = false` every tick, and the event
    handler only sets it true on keydown, never clears it on keyup.
  - If the user dismissed the keyboard (e.g. the system back gesture)
    without closing the shell, there was no way to bring it back short of
    closing and reopening the shell itself. `Shell.Draw()` now calls
    `ShowKeyboard()` again on a fresh tap (edge-detected, not every frame
    of a hold) within the console's own drawn area.
- **App relaunch after an in-game quit.** Using the in-game quit (e.g. the
  Level Editor's File menu) would leave the app crashing silently in the
  background on the next launch -- it'd show the usual debug-build splash
  and then just die, with nothing informative in `adb logcat`, until the
  whole task was swiped away in Recents to force a truly fresh process.
  Root cause: returning from `cmd/doodle-android/main_android.go`'s
  `SDL_main()` only ends the JNI call -- Java's `SDLMain.run()` then calls
  `Activity.finish()`, which ends the *Activity* but not necessarily the
  underlying Android *process*, which Android often keeps alive
  in the background for a fast relaunch. Since `dlopen()` (what
  `System.loadLibrary()` does under the hood) doesn't re-run a shared
  library's init code on a second load within the same process, a
  relaunched Activity in that surviving process would call the *same*
  exported `SDL_main` a second time against a Go runtime and SDL2 C-level
  global state that already went through one full run and was never
  designed to be re-entered. Fixed with an explicit `os.Exit(0)` right
  after `game.Run()` returns, guaranteeing the process actually ends and
  the next launch always starts from a real clean slate.
- **Screen size / editor UX.** The level editor's toolbox and menus were
  designed for a mouse and a desktop-sized window; nothing has been tuned
  yet for small touchscreens beyond the existing touch-vs-mouse cursor
  detection.
- **Gamepad rumble/haptics** aren't wired up (the desktop build doesn't use
  them either, so there's nothing Android-specific broken here, just
  nothing implemented).
- Only lightly tested: this is a build scaffold, not a QA pass. Expect to
  find and fix real bugs the first few times you actually play through it
  on a device.

## Troubleshooting

- **"App may not work properly" 16 KB page size warning** (Android Studio's
  install dialog, or `adb install`'s own compatibility check) -- newer
  Android versions/devices are migrating to a 16 KB memory page size for
  performance, and starting around late 2025 Google Play requires apps
  with native code to support it; devices that already use 16 KB pages
  will show this warning (or in stricter cases, fail to load the library)
  if any bundled `.so` isn't built for it. This was actually happening
  here, confirmed with `llvm-readelf -l some.so | grep LOAD` (look at the
  `Align` column: `0x1000` is 4 KB, `0x4000` is the 16 KB alignment that's
  needed) -- two separate things had to be fixed:
  1. **NDK version.** NDK r27+ defaults its linker to 16 KB-aligned ELF
     segments (older NDKs don't, with no extra flag needed to opt in --
     this project used to pin `26.3.11579264`, discovered to be
     insufficient this way). `app/build.gradle`'s `ndkVersion` and
     whatever `ANDROID_NDK_HOME` points at when you run
     `build-native-libs.sh` need to agree, and both need to be r27+.
     `bootstrap.py` always picks the newest installed NDK automatically,
     so this is normally only something to think about if you're setting
     `ANDROID_NDK_HOME` by hand.
  2. **APK packaging.** 16 KB ELF alignment only matters if the OS can
     `mmap()` the library directly out of the APK in the first place,
     which requires it to be stored *uncompressed* there --
     `app/build.gradle`'s `packagingOptions { jniLibs { useLegacyPackaging
     ... } }` controls that, and this project had it backwards for a
     while (`true`, under the mistaken impression that meant "don't
     compress" -- it's the opposite: "legacy" packaging is the old
     pre-API-23 compressed/extract-at-install-time behavior). It's
     `false` now, matching AGP's own default; if you ever "fix" it back
     the wrong way, this whole class of problem returns, silently, even
     with a new-enough NDK -- worth verifying directly rather than
     trusting the setting, with the SDK's own `zipalign` (in
     `$ANDROID_HOME/build-tools/<version>/`) in check mode:
     ```sh
     zipalign -c -v -P 16 4 app/build/outputs/apk/debug/app-debug.apk
     ```
     Every `.so` should report `(OK)`, ending in `Verification succesful`
     (its own typo, not ours).
- **`Compatibility with CMake < 3.5 has been removed`** while configuring
  SDL_ttf/SDL_mixer's vendored dependencies (e.g. `external/freetype`) --
  those third-party `CMakeLists.txt` files declare an old
  `cmake_minimum_required` that CMake 4.x refuses to honor at all (older
  CMake just warned). `build_sdl_lib()` already passes
  `-DCMAKE_POLICY_VERSION_MINIMUM=3.5` to work around this; if you hit this
  error, make sure your `build-native-libs.sh` includes that flag.
- **`Could NOT find PrivateSDL2 (missing: SDL2_LIBRARY SDL2_INCLUDE_DIR)`**
  while building SDL_ttf or SDL_mixer -- the NDK's CMake toolchain file sets
  `CMAKE_FIND_ROOT_PATH_MODE_LIBRARY`/`INCLUDE`/`PACKAGE` to `ONLY`, which
  restricts `find_library`/`find_path`/`find_package` to only look inside
  `CMAKE_FIND_ROOT_PATH` (by default just the NDK's own sysroot) --
  `CMAKE_PREFIX_PATH` alone doesn't override that restriction, so SDL_ttf
  can't see the SDL2 we just built and installed to `third_party/install/<abi>`.
  `build_sdl_lib()` already works around this by also passing
  `-DCMAKE_FIND_ROOT_PATH="$install_prefix"` (appended to, not replacing,
  the NDK's own path) and setting `CMAKE_FIND_ROOT_PATH_MODE_LIBRARY`/
  `INCLUDE` to `BOTH` (deliberately leaving `PACKAGE` at `ONLY` -- see the
  next entry for why). If you hit this, make sure you're on a version of
  `build-native-libs.sh` that includes those flags.
- **`ld.lld: error: /usr/lib/libOpenGL.so is incompatible with aarch64linux`**
  or **`ld.lld: error: undefined symbol: main`** while building SDL_ttf's
  bundled `glfont`/`showfont` sample programs -- two compounding bugs, both
  now fixed in `build-native-libs.sh`:
  1. SDL_ttf and SDL_mixer's CMake options are prefixed `SDL2TTF_`/
     `SDL2MIXER_`, not `SDLTTF_`/`SDLMIXER_`. CMake silently ignores unknown
     `-D` cache variables instead of erroring, so an earlier version of this
     script that used the wrong prefix quietly fell back to upstream's
     defaults -- including `SDL2TTF_SAMPLES`/`SDL2MIXER_SAMPLES` defaulting
     *on*, which builds example programs we don't need or want.
  2. Those sample programs are desktop demos (one wants real OpenGL, the
     other doesn't define its own `main`), and setting
     `CMAKE_FIND_ROOT_PATH_MODE_PACKAGE` to `BOTH` (an earlier, overly broad
     fix for the `PrivateSDL2` error above) let `find_package(OpenGL)` find
     the *host's* desktop OpenGL and try to link an aarch64 binary against
     an x86_64 `.so`.

  If you hit either error, make sure `build-native-libs.sh` passes
  `-DSDL2TTF_SAMPLES=OFF`/`-DSDL2MIXER_SAMPLES=OFF` and that
  `CMAKE_FIND_ROOT_PATH_MODE_PACKAGE` is `ONLY`, then delete
  `third_party/SDL_ttf/build-<abi>` and `third_party/SDL_mixer/build-<abi>`
  to force a clean reconfigure (a stale `CMakeCache.txt` from a previous
  run can otherwise keep the old, wrong option values).
- **`UnsatisfiedLinkError: dlopen failed: library "libSDL2.so" not found`**
  -- `build-native-libs.sh` didn't run, or didn't run for the ABI of the
  device/emulator you're testing on. Check `app/src/main/jniLibs/<abi>/`
  actually has all five `.so` files for that ABI.
- **`UnsatisfiedLinkError: libc++_shared.so not found`** -- same as above;
  this file comes from the NDK itself, not from the SDL2 build, and is easy
  to forget if you hand-roll the build differently from the provided script.
- **`expected NDK clang at .../bin/aarch64-linux-android21-clang`** from
  `build-native-libs.sh` -- your `ANDROID_NDK_HOME` doesn't match the
  `ANDROID_API` the script is targeting (default 21). Either install that
  API level's platform or override `ANDROID_API` in your environment (it
  must stay `>=` `app/build.gradle`'s `minSdkVersion`).
- **Gradle can't find the SDK** -- create `android/local.properties` from
  `local.properties.example`, or let Android Studio generate it.
- **cmake/ninja not found** -- install both via Android Studio's SDK
  Manager (SDK Tools tab) or your OS package manager, and make sure
  `cmake`/`ninja` are on `PATH`.

## Updating SDL2 versions

The pinned tags live at the top of `scripts/fetch-sdl-sources.sh`
(`SDL_VERSION`, `SDL_TTF_VERSION`, `SDL_MIXER_VERSION`). Bump them, delete
`third_party/` to force a clean re-clone, and re-run both scripts. Sanity
check the new SDL2 C API version is still compatible with whatever version
of `github.com/veandco/go-sdl2` is pinned in `../go.mod` -- go-sdl2's own
README notes which SDL2 releases it's tested against.
