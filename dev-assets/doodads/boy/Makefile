ALL: build

.PHONY: build
build:
	doodad convert -t "Boy" stand-right.png stand-left.png \
		walk-right-1.png walk-right-2.png walk-right-3.png \
		walk-left-1.png walk-left-2.png walk-left-3.png \
		boy.doodad
	doodad install-script boy.js boy.doodad

	doodad edit-doodad --tag "category=creatures" boy.doodad

	cp *.doodad ../../../assets/doodads/