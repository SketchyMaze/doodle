SHELL = /bin/bash

ALL: build

.PHONY: build
build:
	# Build the four basic trapdoors.
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

	# Tag the category for these doodads
	ls -hal *.doodad
	for i in *.doodad; do\
		echo $${i};\
		doodad edit-doodad -q --tag "category=doors" $${i};\
	done

	# Build the Electric Trapdoor.
	doodad convert -t "Electric Trapdoor" electric{1,2,3,4}.png electric-trapdoor.doodad
	doodad install-script electric-trapdoor.js electric-trapdoor.doodad
	doodad edit-doodad -q --tag "category=doors,gizmos" electric-trapdoor.doodad

	cp *.doodad ../../../assets/doodads/
