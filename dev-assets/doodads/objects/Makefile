ALL: build

.PHONY: build
build:
	# Start Flag
	doodad convert -t "Start Flag" start-flag.png start-flag.doodad
	doodad install-script start-flag.js start-flag.doodad

	# Exit Flag
	doodad convert -t "Exit Flag" exit-flag.png exit-flag.doodad
	doodad install-script exit-flag.js exit-flag.doodad

	# Checkpoint Flag
	doodad convert -t "Checkpoint Flag" checkpoint-active.png \
		checkpoint-inactive.png checkpoint-flag.doodad
	doodad install-script checkpoint-flag.js checkpoint-flag.doodad

	# Anvil
	doodad convert -t "Anvil" anvil.png anvil.doodad
	doodad install-script anvil.js anvil.doodad

	for i in *.doodad; do\
		doodad edit-doodad --tag "category=objects" $${i};\
	done
	cp *.doodad ../../../assets/doodads/