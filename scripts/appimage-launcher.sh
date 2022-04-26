#!/bin/bash

cd "$(dirname "$0")"
echo AppImage Sketchy Maze launcher

if [[ "$1" == "doodad" ]]; then
    exec $@;
fi

exec ./sketchymaze $@
