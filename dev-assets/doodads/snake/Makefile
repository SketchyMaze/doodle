ALL: build

.PHONY: build
build:
	doodad convert -t "Snake" left-1.png left-2.png left-3.png right-1.png right-2.png right-3.png \
		attack-left-1.png attack-left-2.png attack-left-3.png attack-right-1.png attack-right-2.png \
		attack-right-3.png snake.doodad
	doodad install-script snake.js snake.doodad

	# Tag the category for these doodads
	for i in *.doodad; do\
		doodad edit-doodad --tag "category=creatures" $${i};\
	done

	cp *.doodad ../../../assets/doodads/