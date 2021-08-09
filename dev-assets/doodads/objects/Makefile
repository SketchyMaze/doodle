ALL: build

.PHONY: build
build:
	doodad convert -t "Exit Flag" exit-flag.png exit-flag.doodad
	doodad install-script exit-flag.js exit-flag.doodad

	doodad convert -t "Start Flag" start-flag.png start-flag.doodad

	for i in *.doodad; do\
		doodad edit-doodad --tag "category=objects" $${i};\
	done
	cp *.doodad ../../../assets/doodads/