ALL: build

.PHONY: build
build:
	doodad convert -t "Crumbly Floor" floor.png shake1.png shake2.png \
		fall1.png fall2.png fall3.png fall4.png fallen.png \
		crumbly-floor.doodad
	doodad install-script crumbly-floor.js crumbly-floor.doodad
	for i in *.doodad; do\
		doodad edit-doodad --tag "category=objects" $${i};\
	done
	cp *.doodad ../../../assets/doodads/