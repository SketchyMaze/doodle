#!/usr/bin/env python3

"""
Bootstrap a Linux or macOS host for building the SketchyMaze Android app.

Installs what it can via your system package manager (cmake, ninja, git, go),
locates Android Studio/SDK/NDK (asking where they are if they're not at the
usual default paths, and remembering the answer for next time), fetches the
SDL2/SDL2_ttf/SDL2_mixer sources the native build needs, and writes an
android/.env file (JAVA_HOME, ANDROID_HOME, ANDROID_NDK_HOME) for the
Makefile -- and for you, via `source .env` -- to use.

Safe to re-run any time: everything it does is idempotent, and once your
paths are known (found at their defaults, or answered once and remembered
in .env) it won't prompt again -- it'll just print a summary of what's in
place and what (if anything) still needs manual attention, like installing
Android Studio itself or an SDK platform via its SDK Manager GUI, which this
script can detect but can't do for you.

See android/README.md's "Quick Start" section for the make targets that
build on top of this (make debug, make debug-install, etc).
"""

import argparse
import platform
import re
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent  # android/
ENV_FILE = ROOT / ".env"
FETCH_SCRIPT = ROOT / "scripts" / "fetch-sdl-sources.sh"
THIRD_PARTY = ROOT / "third_party"

IS_MACOS = platform.system() == "Darwin"

# What each of the *installable* dependencies are called in each package
# manager. Android Studio/SDK/NDK are not in here: no distro packages them
# usefully, so they're handled separately (detected, or asked about).
dep_fedora = ["cmake", "ninja-build", "git", "golang"]
dep_debian = ["cmake", "ninja-build", "git", "golang-go"]
dep_arch = ["cmake", "ninja", "git", "go"]
dep_macos = ["cmake", "ninja", "git", "go"]

# Default install locations. Linux defaults match what a manual "extract the
# tarball to your home directory" Android Studio install looks like (and
# this repo's own README instructions); macOS defaults match where the
# official .dmg installer actually puts things, which is different enough
# that reusing the Linux defaults there would just never match.
if IS_MACOS:
    DEFAULT_STUDIO = Path("/Applications/Android Studio.app")
    DEFAULT_SDK = Path.home() / "Library" / "Android" / "sdk"
else:
    DEFAULT_STUDIO = Path.home() / "android-studio"
    DEFAULT_SDK = Path.home() / "Android" / "Sdk"


class Report:
    """Collects (name, ok, detail, fix) rows for the final summary."""

    def __init__(self):
        self.rows = []

    def ok(self, name, detail=""):
        self.rows.append((name, True, detail, None))

    def fail(self, name, detail, fix):
        self.rows.append((name, False, detail, fix))

    def print_summary(self):
        width = max(len(name) for name, *_ in self.rows)
        print("\n" + "=" * 70)
        print(" SketchyMaze Android bootstrap summary")
        print("=" * 70)
        missing = 0
        for name, ok, detail, fix in self.rows:
            tag = " OK " if ok else "MISS"
            print(f"[{tag}] {name:<{width}}  {detail}")
            if not ok:
                missing += 1
                print(f"         -> {fix}")
        print("=" * 70)
        if missing:
            print(f"{missing} item(s) need manual attention -- see the -> lines above.\n")
        else:
            print("Everything's in place. Try `make debug-install` next.\n")
        return missing


def shell(cmd, **kwargs):
    """Echo and run a shell command, returning its exit code."""
    print("$", " ".join(cmd) if isinstance(cmd, list) else cmd)
    return subprocess.call(cmd, cwd=ROOT, **kwargs)


def which(name):
    return shutil.which(name)


def read_env_file():
    """Parse a previously-written .env (plain `export KEY=value` lines) into a dict."""
    env = {}
    if not ENV_FILE.exists():
        return env
    for line in ENV_FILE.read_text().splitlines():
        line = line.strip()
        if not line or not line.startswith("export "):
            continue
        line = line[len("export "):]
        if "=" not in line:
            continue
        key, _, value = line.partition("=")
        env[key.strip()] = value.strip()
    return env


def install_packages(args, report):
    """Install what's available via the host's package manager."""
    fast = " -y" if args.fast else ""

    # Check IS_MACOS (from platform.system()) before any Linux tool
    # detection: some macOS setups have an `apt` shim floating around,
    # which would otherwise misdetect as Debian.
    if IS_MACOS and which("brew"):
        pkgs = " ".join(dep_macos)
        rc = shell(f"brew install {pkgs}", shell=True)
    elif which("dnf"):
        pkgs = " ".join(dep_fedora)
        rc = shell(f"sudo dnf install{fast} {pkgs}", shell=True)
    elif which("apt"):
        pkgs = " ".join(dep_debian)
        rc = shell(f"sudo apt update && sudo apt install{fast} {pkgs}", shell=True)
    elif which("pacman"):
        pkgs = " ".join(dep_arch)
        rc = shell(f"sudo pacman -S --needed{fast} {pkgs}", shell=True)
    else:
        report.fail(
            "System packages",
            "no supported package manager found (dnf/apt/pacman/brew)",
            f"install these yourself and make sure they're on PATH: {', '.join(dep_debian)}",
        )
        return

    if rc == 0:
        report.ok("System packages", "cmake, ninja, git, go")
    else:
        report.fail(
            "System packages",
            "package manager install command failed (see output above)",
            "re-run this script, or install cmake/ninja/git/go by hand",
        )


def ask_path(args, prompt_text, saved, default, report_name, report, install_hint):
    """
    Resolve a directory path: reuse a still-valid saved answer, else accept
    the default if it exists, else prompt interactively (or, in --fast mode,
    just report it missing without blocking on input).
    """
    if saved and Path(saved).is_dir():
        return Path(saved)

    if default.is_dir():
        return default

    if args.fast:
        report.fail(report_name, f"not found at default ({default})", install_hint)
        return None

    answer = input(f"{prompt_text} [{default}]: ").strip()
    path = Path(answer).expanduser() if answer else default
    if path.is_dir():
        return path

    report.fail(report_name, f"'{path}' doesn't exist", install_hint)
    return None


def find_ndk(sdk, report):
    if sdk is None:
        report.fail("Android NDK", "can't look for it without a valid SDK path", "install the Android SDK first")
        return None

    ndk_root = sdk / "ndk"
    if not ndk_root.is_dir():
        report.fail(
            "Android NDK",
            f"no {ndk_root} folder",
            "Android Studio > More Actions/Tools > SDK Manager > SDK Tools tab > "
            'check "NDK (Side by side)" > Apply',
        )
        return None

    def version_key(p):
        return tuple(int(x) if x.isdigit() else 0 for x in p.name.split("."))

    versions = sorted((p for p in ndk_root.iterdir() if p.is_dir()), key=version_key)
    if not versions:
        report.fail(
            "Android NDK",
            f"{ndk_root} exists but has no versions installed",
            "Android Studio > SDK Manager > SDK Tools tab > check \"NDK (Side by side)\" > Apply",
        )
        return None

    newest = versions[-1]
    detail = newest.name
    if len(versions) > 1:
        detail += f" (newest of {len(versions)} installed; others: {', '.join(v.name for v in versions[:-1])})"
    report.ok("Android NDK", detail)
    return newest


def jbr_home(studio):
    """Where Android Studio's bundled JDK (JBR) lives, which differs on macOS."""
    if IS_MACOS:
        return studio / "Contents" / "jbr" / "Contents" / "Home"
    return studio / "jbr"


def find_java_home(studio, report):
    if studio is None:
        report.fail("JAVA_HOME", "can't derive it without Android Studio", "install Android Studio first")
        return None

    home = jbr_home(studio)
    java_bin = home / "bin" / "java"
    if not java_bin.exists():
        report.fail(
            "JAVA_HOME",
            f"no bundled JDK found at {home}",
            "install a JDK yourself and set JAVA_HOME in android/.env manually",
        )
        return None

    report.ok("JAVA_HOME", str(home))
    return home


def check_go(report):
    if which("go"):
        version = subprocess.run(["go", "version"], capture_output=True, text=True).stdout.strip()
        report.ok("Go toolchain", version)
    else:
        report.fail("Go toolchain", "`go` not on PATH", "install Go (should have been handled above; check PATH)")


def check_sdk_platform(sdk, report):
    """Gradle's compileSdk needs the matching SDK Platform package installed."""
    if sdk is None:
        return
    # Keep this loosely in sync with app/build.gradle's compileSdk.
    wanted = "34"
    target = sdk / "platforms" / f"android-{wanted}"
    if target.is_dir():
        report.ok(f"SDK Platform {wanted}", str(target))
    else:
        report.fail(
            f"SDK Platform {wanted}",
            f"no {target}",
            f"Android Studio > SDK Manager > SDK Platforms tab > check the API {wanted} row > Apply",
        )


def check_gradle_wrapper(report):
    gradlew = ROOT / "gradlew"
    if gradlew.exists():
        report.ok("Gradle wrapper", str(gradlew))
    else:
        report.fail(
            "Gradle wrapper",
            "android/gradlew doesn't exist yet",
            'open android/ in Android Studio once (it generates the wrapper automatically), '
            "or if you have a system Gradle: cd android && gradle wrapper --gradle-version 8.7",
        )


def sdl_pinned_version():
    """Read the pinned SDL_VERSION out of fetch-sdl-sources.sh, so this doesn't drift out of sync."""
    text = FETCH_SCRIPT.read_text()
    m = re.search(r'^SDL_VERSION=(\S+)', text, re.MULTILINE)
    return m.group(1) if m else "unknown version"


def fetch_sdl(args, report):
    sdl_dir = THIRD_PARTY / "SDL"

    if args.refetch:
        for name in ("SDL", "SDL_ttf", "SDL_mixer"):
            d = THIRD_PARTY / name
            if d.is_dir():
                print(f"Removing {d} to force a fresh fetch...")
                shutil.rmtree(d)

    if sdl_dir.is_dir():
        report.ok("SDL2 sources", f"already fetched at {sdl_dir} (pass --refetch to update)")
        return

    print("\nFetching SDL2/SDL2_ttf/SDL2_mixer sources -- this clones a few repos "
          "and their submodules, may take a minute...\n")
    rc = shell([str(FETCH_SCRIPT)])
    if rc == 0:
        report.ok("SDL2 sources", f"fetched {sdl_pinned_version()}")
    else:
        report.fail(
            "SDL2 sources",
            "fetch-sdl-sources.sh failed (see output above)",
            "check your network connection and re-run: ./scripts/fetch-sdl-sources.sh",
        )


def write_env_file(studio, java_home, sdk, ndk):
    lines = ["# Generated by bootstrap.py -- re-run it to refresh this, don't hand-edit.",
             "# Usable both as `source android/.env` and via the Makefile's `include .env`.",
             "# ANDROID_STUDIO_HOME is only here so re-running bootstrap.py doesn't have to",
             "# ask where Android Studio is again; nothing else reads it."]
    if studio:
        lines.append(f"export ANDROID_STUDIO_HOME={studio}")
    if java_home:
        lines.append(f"export JAVA_HOME={java_home}")
    if sdk:
        lines.append(f"export ANDROID_HOME={sdk}")
        lines.append(f"export ANDROID_SDK_ROOT={sdk}")
    if ndk:
        lines.append(f"export ANDROID_NDK_HOME={ndk}")
    ENV_FILE.write_text("\n".join(lines) + "\n")
    print(f"\nWrote {ENV_FILE}")


def main(args):
    print(__doc__)
    if not args.fast:
        input("Press Enter to begin (or re-run with --fast to skip this and any other prompts).")

    report = Report()
    saved = read_env_file()

    install_packages(args, report)

    studio = ask_path(
        args,
        "Path to your Android Studio install",
        saved.get("ANDROID_STUDIO_HOME"),
        DEFAULT_STUDIO,
        "Android Studio",
        report,
        "download and install it from https://developer.android.com/studio",
    )
    if studio:
        report.ok("Android Studio", str(studio))

    sdk = ask_path(
        args,
        "Path to your Android SDK",
        saved.get("ANDROID_HOME"),
        DEFAULT_SDK,
        "Android SDK",
        report,
        "open Android Studio once (More Actions > SDK Manager) to have it install the SDK, "
        "or point this at wherever you already have one",
    )
    if sdk:
        report.ok("Android SDK", str(sdk))

    ndk = find_ndk(sdk, report)
    java_home = find_java_home(studio, report)
    check_go(report)
    check_sdk_platform(sdk, report)
    check_gradle_wrapper(report)
    fetch_sdl(args, report)

    write_env_file(studio, java_home, sdk, ndk)

    missing = report.print_summary()
    sys.exit(1 if missing else 0)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Bootstrap Android build dependencies for SketchyMaze")
    parser.add_argument("--fast", "-f",
        action="store_true",
        help="Non-interactive: accept defaults and package manager prompts automatically instead of asking",
    )
    parser.add_argument("--refetch",
        action="store_true",
        help="Re-fetch SDL2/SDL2_ttf/SDL2_mixer sources even if already present (e.g. to pick up a version bump)",
    )
    main(parser.parse_args())
