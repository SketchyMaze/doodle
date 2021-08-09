ALL: build

.PHONY: build
build:
	# Start Flag
	doodad convert -t "Start Flag" start-flag.png start-flag.doodad

	# Exit Flag
	doodad convert -t "Exit Flag" exit-flag.png exit-flag.doodad
	doodad install-script exit-flag.js exit-flag.doodad

	# Anvil
	doodad convert -t "Anvil" anvil.png anvil.doodad
	doodad install-script anvil.js anvil.doodad

	for i in *.doodad; do\
		doodad edit-doodad --tag "category=objects" $${i};\
	done
	cp *.doodad ../../../assets/doodads/