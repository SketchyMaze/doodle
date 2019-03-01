#!/bin/bash

# Run this script from the root of the repo.
if [[ ! -f "./docker/dist-ubuntu.sh" ]]; then
	echo "Run this script from the root of the doodle repo, i.e.: ./docker/dist-ubuntu.sh"
	exit 1
fi

sudo docker build -t doodle_ubuntu -f ./docker/Ubuntu.dockerfile .
sudo docker run --rm -v "$(pwd)/docker/ubuntu:/mnt/export:z" doodle_ubuntu
