#!/bin/bash

# Build all the doodads from their source files.
if [[ ! -d "./azulian" ]]; then
	echo Run this script from the dev-assets/doodads/ working directory.
	exit 1
fi

mkdir -p ../../assets/doodads

buttons() {
	cd buttons/

	doodad convert -t "Sticky Button" sticky1.png sticky2.png sticky-button.doodad
	doodad install-script sticky.js sticky-button.doodad
	cp sticky-button.doodad ../../../assets/doodads/

	doodad convert -t "Button" button1.png button2.png button.doodad
	doodad install-script button.js button.doodad
	cp button.doodad ../../../assets/doodads/

	doodad convert -t "Button Type B" typeB1.png typeB2.png button-typeB.doodad
	doodad install-script button.js button-typeB.doodad
	cp button-typeB.doodad ../../../assets/doodads/

	cd ..
}

azulians() {
	cd azulian/

	doodad convert -t "Blue Azulian" blu-front.png blu-back.png \
		blu-wr{1,2,3,4}.png blu-wl{1,2,3,4}.png azu-blu.doodad
	doodad install-script azulian.js azu-blu.doodad
	cp azu-blu.doodad ../../../assets/doodads/

	cd ..
}

buttons
azulians
