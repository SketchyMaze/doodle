#!/usr/bin/env python

"""
Full setup, from scratch to distribution, of Project: Doodle.

Run this script from an empty working directory. All git repos will be cloned
here (or updated if already existing) and the app will be fully built including
fonts, default levels and doodads, sound effects and music for your current
system. Useful to quickly bootstrap a build on weird operating systems like
macOS or Linux on ARM (Pinephone).

First ensure that your SSH key is authorized on git.kirsle.net to download
the repos easily. This script will also handle installing the SDL2 dependencies
on Fedora, Debian and macOS type systems.
"""

import sys
import os
import os.path
import subprocess
import pathlib

# Git repositories.
repos = {
    "git@git.kirsle.net:apps/doodle-masters": "masters",
    "git@git.kirsle.net:apps/doodle-vendor": "vendor",
    "git@git.kirsle.net:apps/doodle-rtp": "rtp",
    "git@git.kirsle.net:go/render": "render",
    "git@git.kirsle.net:go/ui": "ui",
    "git@git.kirsle.net:go/audio": "audio",
}
repos_github = {
    # GitHub mirrors of the above.
    "git@github.com:kirsle/render": "render",
    "git@github.com:kirsle/ui": "ui",
    "git@github.com:kirsle/audio": "audio",
    # TODO: the rest
}

# Software dependencies.
dep_fedora = ["make", "golang", "SDL2-devel", "SDL2_ttf-devel", "SDL2_mixer-devel"]
dep_debian = ["make", "golang", "libsdl2-dev", "libsdl2-ttf-dev", "libsdl2-mixer-dev"]
dep_macos = ["golang", "sdl2", "sdl2_ttf", "sdl2_mixer", "pkg-config"]
dep_arch = ["go", "sdl2", "sdl2_ttf", "sdl2_mixer"]


# Absolute path to current working directory.
ROOT = pathlib.Path().absolute()


def main():
    print(
        "Project: Doodle Full Installer\n"
        "Current working directory: {root}\n"
        "Ensure your SSH keys are set up on git.kirsle.net to easily clone repos.\n"
        "Also check your $GOPATH is set and your $PATH will run binaries installed,\n"
        "for e.g. GOPATH=$HOME/go and PATH includes $HOME/go/bin; otherwise the\n"
        "'make doodads' command won't function later."
        .format(root=ROOT)
    )
    input("Press Enter to begin.")
    install_deps()
    clone_repos()
    patch_gomod()
    copy_assets()
    install_doodad()
    build()


def install_deps():
    """Install system dependencies."""
    if shell("which rpm") == 0 and shell("which dnf") == 0:
        # Fedora-like.
        if shell("rpm -q {}".format(' '.join(dep_fedora))) != 0:
            must_shell("sudo dnf install {}".format(' '.join(dep_fedora)))
    elif shell("which brew") == 0:
        # MacOS, as Catalina has an apt command now??
        must_shell("brew install {}".format(' '.join(dep_macos)))
    elif shell("which apt") == 0:
        # Debian-like.
        if shell("dpkg-query -l {}".format(' '.join(dep_debian))) != 0:
            must_shell("sudo apt update && sudo apt install {}".format(' '.join(dep_debian)))
    elif shell("which pacman") == 0:
        # Arch-like.
        must_shell("sudo pacman -S {}".format(' '.join(dep_arch)))
    else:
        print("Warning: didn't detect your package manager to install SDL2 and other dependencies")



def clone_repos():
    """Clone or update all the git repos"""
    if not os.path.isdir("./deps"):
        os.mkdir("./deps")
    os.chdir("./deps")
    for url, name in repos.items():
        if os.path.isdir(name):
            os.chdir(name)
            must_shell("git pull --ff-only")
            os.chdir("..")
        else:
            must_shell("git clone {} {}".format(url, name))
    os.chdir("..")  # back to doodle root


def patch_gomod():
    """Patch the doodle/go.mod to use local paths to other repos."""
    if shell("grep -e 'replace git.kirsle.net' go.mod") != 0:
        with open("go.mod", "a") as fh:
            fh.write(
                "\n\nreplace git.kirsle.net/go/render => {root}/deps/render\n"
                "replace git.kirsle.net/go/ui => {root}/deps/ui\n"
                "replace git.kirsle.net/go/audio => {root}/deps/audio\n"
                .format(root=ROOT)
            )


def copy_assets():
    """Copy assets from other repos into doodle."""
    if not os.path.isdir("assets/fonts"):
        shell("cp -rv deps/vendor/fonts assets/fonts")
    if not os.path.isdir("assets/levelpacks"):
        shell("cp -rv deps/masters/levelpacks/levelpacks assets/levelpacks")
    if not os.path.isdir("rtp"):
        shell("mkdir -p rtp && cp -rv deps/rtp/* rtp/")


def install_doodad():
    """Install the doodad CLI tool from the doodle repo."""
    must_shell("go install git.kirsle.net/apps/doodle/cmd/doodad")


def build():
    """Build the game."""
    must_shell("make dist")


def shell(cmd):
    """Echo and run a shell command"""
    print("$ ", cmd)
    return subprocess.call(cmd, shell=True)


def must_shell(cmd):
    """Run a shell command which MUST succeed."""
    assert shell(cmd) == 0


if __name__ == "__main__":
    main()
