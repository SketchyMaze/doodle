#!/bin/bash

# Run this script from the root of the repo.
if [[ ! -f "./docker/dist-fedora.sh" ]]; then
	echo "Run this script from the root of the doodle repo, i.e.: ./docker/dist-fedora.sh"
	exit 1
fi

sudo docker build -t doodle_fedora -f ./docker/Fedora.dockerfile .
sudo docker run --rm -v "$(pwd)/docker/fedora:/mnt/export:z" doodle_fedora
