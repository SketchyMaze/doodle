ALL: build

.PHONY: build
build:
	doodad convert -t "Switch" switch-off.png switch-on.png switch.doodad
	doodad convert -t "Floor Switch" down-off.png down-on.png switch-down.doodad
	doodad convert -t "Left Switch" left-off.png left-on.png switch-left.doodad
	doodad convert -t "Right Switch" right-off.png right-on.png switch-right.doodad

	doodad install-script switch.js switch.doodad
	doodad install-script switch.js switch-down.doodad
	doodad install-script switch.js switch-left.doodad
	doodad install-script switch.js switch-right.doodad

	# Tag the category for these doodads
	for i in *.doodad; do\
		doodad edit-doodad --tag "category=gizmos" $${i};\
	done

	cp *.doodad ../../../assets/doodads/