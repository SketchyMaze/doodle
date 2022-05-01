ALL: build

.PHONY: build
build:
	doodad convert -t "Crusher" sleep.png peek-left.png peek-right.png \
		angry.png ouch.png crusher.doodad
	doodad install-script crusher.js crusher.doodad

	# Tag the category for these doodads
	for i in *.doodad; do\
		doodad edit-doodad --tag "category=creatures" $${i};\
	done

	cp *.doodad ../../../assets/doodads/