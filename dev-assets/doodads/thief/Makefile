ALL: build

.PHONY: build
build:
	doodad convert -t "Thief" stand-right.png stand-left.png \
		thief.doodad
	doodad install-script thief.js thief.doodad

	doodad edit-doodad --tag "category=creatures" thief.doodad

	cp *.doodad ../../../assets/doodads/