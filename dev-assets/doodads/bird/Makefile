ALL: build

.PHONY: build
build:
	doodad convert -t "Bird (red)" left-1.png left-2.png right-1.png right-2.png \
		dive-left.png dive-right.png bird-red.doodad
	doodad install-script bird.js bird-red.doodad

	# Tag the category for these doodads
	for i in *.doodad; do\
		doodad edit-doodad --tag "category=creatures" $${i};\
	done

	cp *.doodad ../../../assets/doodads/