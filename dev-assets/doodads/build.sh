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

switches() {
	cd switches/

	doodad convert -t "Switch" switch-off.png switch-on.png switch.doodad
	doodad convert -t "Floor Switch" down-off.png down-on.png switch-down.doodad
	doodad convert -t "Left Switch" left-off.png left-on.png switch-left.doodad
	doodad convert -t "Right Switch" right-off.png right-on.png switch-right.doodad

	doodad install-script switch.js switch.doodad
	doodad install-script switch.js switch-down.doodad
	doodad install-script switch.js switch-left.doodad
	doodad install-script switch.js switch-right.doodad

	cp *.doodad ../../../assets/doodads/
	cd ..
}

doors() {
	cd doors/
	./build.sh
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

	doodad edit-doodad -q --tag direction=down trapdoor-down.doodad
	doodad edit-doodad -q --tag direction=left trapdoor-left.doodad
	doodad edit-doodad -q --tag direction=right trapdoor-right.doodad
	doodad edit-doodad -q --tag direction=up trapdoor-up.doodad

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

	doodad convert -t "Start Flag" start-flag.png start-flag.doodad

	cp *.doodad ../../../assets/doodads/

	cd ../crumbly-floor

	doodad convert -t "Crumbly Floor" floor.png shake1.png shake2.png \
		fall1.png fall2.png fall3.png fall4.png fallen.png \
		crumbly-floor.doodad
	doodad install-script crumbly-floor.js crumbly-floor.doodad
	cp *.doodad ../../../assets/doodads/

	cd ..
}

onoff() {
	cd on-off/

	doodad convert -t "State Button" blue-button.png orange-button.png state-button.doodad
	doodad install-script state-button.js state-button.doodad

	doodad convert -t "State Block (Blue)" blue-on.png blue-off.png state-block-blue.doodad
	doodad install-script state-block-blue.js state-block-blue.doodad

	doodad convert -t "State Block (Orange)" orange-off.png orange-on.png state-block-orange.doodad
	doodad install-script state-block-orange.js state-block-orange.doodad

	cp *.doodad ../../../assets/doodads/

	cd ..
}

buttons
switches
doors
trapdoors
azulians
objects
onoff
doodad edit-doodad -quiet -lock -author "Noah" ../../assets/doodads/*.doodad
doodad edit-doodad -hide ../../assets/doodads/azu-blu.doodad
