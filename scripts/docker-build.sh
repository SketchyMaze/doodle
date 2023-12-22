#!/bin/bash

# Script to build the Docker CI to export Linux and Windows, 64- and 32-bit releases.

# Ensure we run this from the doodle root.
if [[ ! -f "./cmd/doodle/main.go" ]]; then
    echo "Run this script from the root of your doodle engine checkout."
    echo "This is probably up one directory from where this script lives."
    exit 1
fi

# Fedora: we will need to setenforce permissive to copy the artifacts from the
# Docker container out to the host.
if type "getenforce" > /dev/null; then
    current=`getenforce`;
    if [ $current -eq "Enforcing"]; then
        echo "Your current SELinux policy is set to: $current."
        echo "This will prevent the final built artifacts being moved from the Docker"
        echo "container into the output directory on your host."
        echo ""

        read -e -p "Can I run command 'sudo setenforce permissive'? [yn]" choice
        [[ "$choice" == [Yy]* ]] && sudo setenforce permissive || echo "That was a no"
    fi
fi

# If we don't have podman installed, substitute it for docker.
if ! type "podman" > /dev/null; then
    echo "Note: podman not found, trying docker instead."
    alias podman=`which docker`
fi

mkdir -p docker-artifacts
podman build --cap-add SYS_ADMIN --device /dev/fuse -t doodle_docker .
podman run --rm --mount type=bind,src=$(shell pwd)/docker-artifacts,dst=/mnt/export doodle_docker