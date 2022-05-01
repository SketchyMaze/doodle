ALL: build

.PHONY: build
build:
	###
	# Gemstones
	###

	doodad convert -t "Gemstone (Green)" green-1.png green-2.png green-3.png green-4.png \
		gem-green.doodad
	doodad install-script gemstone.js gem-green.doodad
	doodad edit-doodad --tag "color=green" gem-green.doodad

	doodad convert -t "Gemstone (Red)" red-1.png red-2.png red-3.png red-4.png \
		gem-red.doodad
	doodad install-script gemstone.js gem-red.doodad
	doodad edit-doodad --tag "color=red" gem-red.doodad

	doodad convert -t "Gemstone (Blue)" blue-1.png blue-2.png blue-3.png blue-4.png \
		gem-blue.doodad
	doodad install-script gemstone.js gem-blue.doodad
	doodad edit-doodad --tag "color=blue" gem-blue.doodad

	doodad convert -t "Gemstone (Yellow)" yellow-1.png yellow-2.png yellow-3.png yellow-4.png \
		gem-yellow.doodad
	doodad install-script gemstone.js gem-yellow.doodad
	doodad edit-doodad --tag "color=yellow" gem-yellow.doodad

	###
	# Totems
	###

	doodad convert -t "Gemstone Totem (Green)" totem-green-1.png totem-green-2.png totem-green-3.png \
		totem-green-4.png totem-green-0.png gem-totem-green.doodad
	doodad install-script totem.js gem-totem-green.doodad
	doodad edit-doodad --tag "color=green" gem-totem-green.doodad

	doodad convert -t "Gemstone Totem (Yellow)" totem-yellow-1.png totem-yellow-2.png totem-yellow-3.png \
		totem-yellow-4.png totem-yellow-0.png gem-totem-yellow.doodad
	doodad install-script totem.js gem-totem-yellow.doodad
	doodad edit-doodad --tag "color=yellow" gem-totem-yellow.doodad

	doodad convert -t "Gemstone Totem (Blue)" totem-blue-1.png totem-blue-2.png totem-blue-3.png \
		totem-blue-4.png totem-blue-0.png gem-totem-blue.doodad
	doodad install-script totem.js gem-totem-blue.doodad
	doodad edit-doodad --tag "color=blue" gem-totem-blue.doodad

	doodad convert -t "Gemstone Totem (Red)" totem-red-1.png totem-red-2.png totem-red-3.png \
		totem-red-4.png totem-red-0.png gem-totem-red.doodad
	doodad install-script totem.js gem-totem-red.doodad
	doodad edit-doodad --tag "color=red" gem-totem-red.doodad

	# Tag the category for these doodads
	for i in *.doodad; do\
		doodad edit-doodad --tag "category=doors" $${i};\
	done

	cp *.doodad ../../../assets/doodads/