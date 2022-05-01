ALL: build

.PHONY: build
build:
	doodad convert -t "Bird (red)" red/left-1.png red/left-2.png red/right-1.png \
		red/right-2.png red/dive-left.png red/dive-right.png \
		bird-red.doodad
	doodad install-script bird.js bird-red.doodad
	doodad edit-doodad --tag "color=red" bird-red.doodad

	doodad convert -t "Bird (blue)" blue/left-1.png blue/left-2.png blue/right-1.png \
		blue/right-2.png blue/dive-left.png blue/dive-right.png \
		bird-blue.doodad
	doodad install-script bird.js bird-blue.doodad
	doodad edit-doodad --tag "color=blue" bird-blue.doodad

	# Tag the category for these doodads
	for i in *.doodad; do\
		doodad edit-doodad --tag "category=creatures" $${i};\
	done

	cp *.doodad ../../../assets/doodads/