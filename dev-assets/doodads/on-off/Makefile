ALL: build

.PHONY: build
build:
	doodad convert -t "State Button" blue-button.png orange-button.png state-button.doodad
	doodad install-script state-button.js state-button.doodad

	doodad convert -t "State Block (Blue)" blue-on.png blue-off.png state-block-blue.doodad
	doodad install-script state-block-blue.js state-block-blue.doodad

	doodad convert -t "State Block (Orange)" orange-off.png orange-on.png state-block-orange.doodad
	doodad install-script state-block-orange.js state-block-orange.doodad

	for i in *.doodad; do\
		doodad edit-doodad --tag "category=gizmos" $${i};\
	done

	cp *.doodad ../../../assets/doodads/