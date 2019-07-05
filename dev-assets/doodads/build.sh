#!/bin/bash

# Build all the doodads from their source files.
if [[ ! -d "./azulian" ]]; then
	echo Run this script from the dev-assets/doodads/ working directory.
	exit 1
fi

mkdir -p ../../assets/doodads

buttons() {
	cd buttons/

	doodad convert -t "Sticky Button" sticky1.png sticky2.png button-sticky.doodad
	doodad install-script sticky.js button-sticky.doodad

	doodad convert -t "Button" button1.png button2.png button.doodad
	doodad install-script button.js button.doodad

	doodad convert -t "Button Type B" typeB1.png typeB2.png button-typeB.doodad
	doodad install-script button.js button-typeB.doodad

	cp button*.doodad ../../../assets/doodads/
	cd ..
}

doors() {
	cd doors/

	doodad convert -t "Red Door" red1.png red2.png door-red.doodad
	doodad install-script locked-door.js door-red.doodad

	doodad convert -t "Blue Door" blue1.png blue2.png door-blue.doodad
	doodad install-script locked-door.js door-blue.doodad

	doodad convert -t "Green Door" green1.png green2.png door-green.doodad
	doodad install-script locked-door.js door-green.doodad

	doodad convert -t "Yellow Door" yellow1.png yellow2.png door-yellow.doodad
	doodad install-script locked-door.js door-yellow.doodad

	doodad convert -t "Red Key" red-key.png key-red.doodad
	doodad install-script keys.js key-red.doodad

	doodad convert -t "Blue Key" blue-key.png key-blue.doodad
	doodad install-script keys.js key-blue.doodad

	doodad convert -t "Green Key" green-key.png key-green.doodad
	doodad install-script keys.js key-green.doodad

	doodad convert -t "Yellow Key" yellow-key.png key-yellow.doodad
	doodad install-script keys.js key-yellow.doodad

	doodad convert -t "Electric Door" electric{1,2,3,4}.png door-electric.doodad
	doodad install-script electric-door.js door-electric.doodad

	cp door-*.doodad key-*.doodad ../../../assets/doodads/

	cd ..
}

trapdoors() {
	cd trapdoors/

	doodad convert -t "Trapdoor" down{1,2,3,4}.png trapdoor-down.doodad
	doodad convert -t "Trapdoor Left" left{1,2,3,4}.png trapdoor-left.doodad
	doodad convert -t "Trapdoor Right" right{1,2,3,4}.png trapdoor-right.doodad
	doodad convert -t "Trapdoor Up" up{1,2,3,4}.png trapdoor-up.doodad
	doodad install-script trapdoor.js trapdoor-down.doodad
	doodad install-script trapdoor.js trapdoor-left.doodad
	doodad install-script trapdoor.js trapdoor-right.doodad
	doodad install-script trapdoor.js trapdoor-up.doodad

	cp trapdoor-*.doodad ../../../assets/doodads/

	cd ..
}

azulians() {
	cd azulian/

	doodad convert -t "Blue Azulian" blu-front.png blu-back.png \
		blu-wr{1,2,3,4}.png blu-wl{1,2,3,4}.png azu-blu.doodad
	doodad install-script azulian.js azu-blu.doodad

	doodad convert -t "Red Azulian" red-front.png red-back.png \
		red-wr{1,2,3,4}.png red-wl{1,2,3,4}.png azu-red.doodad
	doodad install-script azulian-red.js azu-red.doodad

	cp azu-*.doodad ../../../assets/doodads/

	cd ..
}

objects() {
	cd objects/

	doodad convert -t "Exit Flag" exit-flag.png exit-flag.doodad
	doodad install-script exit-flag.js exit-flag.doodad

	cp *.doodad ../../../assets/doodads/

	cd ..
}

buttons
doors
trapdoors
azulians
objects
