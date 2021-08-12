#!/bin/bash

# Build all the doodads from their source files.
if [[ ! -d "./azulian" ]]; then
	echo Run this script from the dev-assets/doodads/ working directory.
	exit 1
fi

mkdir -p ../../assets/doodads

boy() {
	cd boy/
	make
	cd ..

	cd thief/
	make
	cd ..
}

buttons() {
	cd buttons/
	make
	cd ..
}

switches() {
	cd switches/
	make
	cd ..
}

doors() {
	cd doors/
	make
	cd ..
}

trapdoors() {
	cd trapdoors/
	make
	cd ..
}

azulians() {
	cd azulian/
	make
	cd ..
}

mobs() {
	cd bird/
	make
	cd ..
}

objects() {
	cd objects/
	make
	cd ..

	cd box/
	make
	cd ..

	cd crumbly-floor/
	make
	cd ..
}

onoff() {
	cd on-off/
	make
	cd ..
}

warpdoor() {
	cd warp-door/
	make
	cd ..
}

boy
buttons
switches
doors
trapdoors
azulians
mobs
objects
onoff
warpdoor
doodad edit-doodad -quiet -lock -author "Noah" ../../assets/doodads/*.doodad
doodad edit-doodad ../../assets/doodads/azu-blu.doodad
doodad edit-doodad -hide ../../assets/doodads/boy.doodad
