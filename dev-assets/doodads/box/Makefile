ALL: build

.PHONY: build
build:
	doodad convert -t "Box" box-1.png box-2.png \
		box-3.png box-4.png box.doodad
	doodad install-script box.js box.doodad

	for i in *.doodad; do \
		doodad edit-doodad --tag "category=objects" $${i}; \
	done

	cp *.doodad ../../../assets/doodads/