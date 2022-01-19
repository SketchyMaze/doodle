SHELL = /bin/bash

ALL: build

.PHONY: build
build:
	doodad convert -t "Blue Azulian" blu-front.png blu-back.png \
		blu-wr{1,2,3,4}.png blu-wl{1,2,3,4}.png azu-blu.doodad
	doodad edit-doodad --tag "color=blue" azu-blu.doodad
	doodad install-script azulian.js azu-blu.doodad

	doodad convert -t "Red Azulian" red-front.png red-back.png \
		red-wr{1,2,3,4}.png red-wl{1,2,3,4}.png azu-red.doodad
	doodad edit-doodad --tag "color=red" azu-red.doodad
	doodad install-script azulian.js azu-red.doodad

	doodad convert -t "White Azulian" white-front.png white-back.png \
		white-wr{1,2,3,4}.png white-wl{1,2,3,4}.png azu-white.doodad
	doodad edit-doodad --tag "color=white" azu-white.doodad
	doodad install-script azulian.js azu-white.doodad

	# Tag the category for these doodads
	for i in *.doodad; do\
		doodad edit-doodad --tag "category=creatures" $${i};\
	done

	cp *.doodad ../../../assets/doodads/
