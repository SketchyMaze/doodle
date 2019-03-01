#!/bin/bash

# Run this script from the root of the repo.
if [[ ! -f "./docker/dist-debian.sh" ]]; then
	echo "Run this script from the root of the doodle repo, i.e.: ./docker/dist-debian.sh"
	exit 1
fi

sudo docker build -t doodle_debian -f ./docker/Debian.dockerfile .
sudo docker run --rm -v "$(pwd)/docker/debian:/mnt/export:z" doodle_debian
