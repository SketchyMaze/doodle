ALL: build

.PHONY: build
build:
	doodad convert -t "Sticky Button" sticky1.png sticky2.png button-sticky.doodad
	doodad install-script sticky.js button-sticky.doodad

	doodad convert -t "Button" button1.png button2.png button.doodad
	doodad install-script button.js button.doodad

	doodad convert -t "Button Type B" typeB1.png typeB2.png button-typeB.doodad
	doodad install-script button.js button-typeB.doodad

	# Tag the category for these doodads
	for i in *.doodad; do\
		doodad edit-doodad --tag "category=gizmos" $${i};\
	done

	cp *.doodad ../../../assets/doodads/