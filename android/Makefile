SHELL := /bin/bash

# Loads JAVA_HOME/ANDROID_HOME/ANDROID_SDK_ROOT/ANDROID_NDK_HOME once
# `make setup` (bootstrap.py) has written them out. The leading "-" means
# Make won't complain if .env doesn't exist yet; these vars then just fall
# through to whatever's already in your shell environment.
-include .env

# Read from app/build.gradle rather than hardcoded, so a future rename (see
# git history for the last one) doesn't also require updating this file.
PACKAGE := $(shell grep -m1 applicationId app/build.gradle | cut -d'"' -f2)
ACTIVITY := $(PACKAGE)/$(PACKAGE).MainActivity

APK_DEBUG := app/build/outputs/apk/debug/app-debug.apk
APK_RELEASE := app/build/outputs/apk/release/app-release.apk

# ------------------------------------------------------------------------
# Setup
# ------------------------------------------------------------------------

# `make setup` installs/checks host dependencies (Android Studio, SDK, NDK,
# cmake, ninja, go, ...) and fetches the SDL2 sources. Safe to re-run --
# see bootstrap.py's own docstring.
.PHONY: setup
setup:
	python3 bootstrap.py

# `make setup-refetch` also re-fetches SDL2/SDL2_ttf/SDL2_mixer even if
# already present, e.g. to pick up a version bump in fetch-sdl-sources.sh.
.PHONY: setup-refetch
setup-refetch:
	python3 bootstrap.py --refetch

# `make doctor` is `make setup` in non-interactive, report-only mode: prints
# what's configured and what's missing without prompting or fetching.
.PHONY: doctor
doctor:
	python3 bootstrap.py --fast

.PHONY: check-gradlew
check-gradlew:
	@if [ ! -x ./gradlew ]; then \
		echo "error: android/gradlew not found." >&2; \
		echo "Run 'make setup' first, or open android/ in Android Studio once to generate it." >&2; \
		exit 1; \
	fi

# ------------------------------------------------------------------------
# Building
# ------------------------------------------------------------------------

# `make native` cross-compiles SDL2/SDL2_ttf/SDL2_mixer and libdoodle.so for
# every ABI (see scripts/build-native-libs.sh). This is the slow step --
# only re-run it after changing Go code, not for Java/Gradle-only changes.
.PHONY: native
native:
	./scripts/build-native-libs.sh

# `make debug` does a full rebuild: native libs + the debug APK. This is
# the "deep" build -- use it after touching any Go code.
.PHONY: debug
debug: check-gradlew native
	./gradlew assembleDebug

# `make debug-install` is the one-command "I changed Go code, get it back
# on my device" loop: full rebuild, force-stop the running app (if any, so
# the reinstall doesn't fight a live process), install, done.
.PHONY: debug-install
debug-install: debug stop
	./gradlew installDebug

# `make apk` / `make apk-install` skip the native rebuild and only redo the
# Java/Gradle/resource layer -- much faster, use these when you're *not*
# touching cmd/doodle-android or any Go code (e.g. iterating on
# MainActivity.java, AndroidManifest.xml, resources).
.PHONY: apk
apk: check-gradlew
	./gradlew assembleDebug

.PHONY: apk-install
apk-install: apk stop
	./gradlew installDebug

# `make build` is `make debug`'s production equivalent: full native rebuild
# + a release APK. Debug-signed by default (see README.md's "Signing a
# release build") so this and `make install` below work out of the box for
# local testing; swap in a real signingConfig before distributing one.
.PHONY: build
build: check-gradlew native
	./gradlew assembleRelease

# `make install` builds (see above) and installs the production APK.
.PHONY: install
install: build stop
	adb install -r $(APK_RELEASE)

# ------------------------------------------------------------------------
# Device helpers
# ------------------------------------------------------------------------

# `make stop` force-stops the app if it's running. The leading "-" means a
# "not currently running" failure here doesn't fail whatever depends on it.
.PHONY: stop
stop:
	-adb shell am force-stop $(PACKAGE)

# `make run` launches the currently-installed build without rebuilding --
# handy after `make stop`, or to relaunch post-crash.
.PHONY: run
run:
	adb shell am start -n $(ACTIVITY)

# `make uninstall` removes the app (and its save data) from the device.
.PHONY: uninstall
uninstall:
	-adb uninstall $(PACKAGE)

# `make log` tails the game's own SDL-tagged log output (rendering,
# lifecycle, and anything logged via SDL's own Log()/LogError() -- NOT
# pkg/log output, which goes to a file on-device instead; see README.md's
# "Assets and save data" section for how to pull that file via adb).
.PHONY: log
log:
	adb logcat -s SDL:V

# ------------------------------------------------------------------------
# Cleanup
# ------------------------------------------------------------------------

# `make clean` removes Gradle's build outputs (APKs, merged manifests, etc).
.PHONY: clean
clean:
	@if [ -x ./gradlew ]; then ./gradlew clean; else rm -rf app/build build; fi

# `make distclean` also removes everything `make setup`/`make native`
# produced: the fetched SDL2 sources, the built .so files, SDL's Java glue,
# and .env -- forcing a completely from-scratch `make setup` next time.
.PHONY: distclean
distclean: clean
	rm -rf third_party
	rm -rf app/src/main/jniLibs/*/*.so
	rm -rf app/src/main/java/org/libsdl/app
	rm -f .env
