ALL: build

.PHONY: build
build:
	doodad convert -t "Warp Door" door-1.png door-2.png door-3.png door-4.png warp-door.doodad
	doodad edit-doodad -q --tag color=none warp-door.doodad
	doodad install-script warp-door.js warp-door.doodad

	doodad convert -t "Warp Door (Blue)" blue-1.png blue-2.png blue-3.png blue-4.png blue-off.png \
		warp-door-blue.doodad
	doodad edit-doodad -q --tag color=blue warp-door-blue.doodad
	doodad install-script warp-door.js warp-door-blue.doodad

	doodad convert -t "Warp Door (Orange)" orange-off.png orange-1.png orange-2.png orange-3.png orange-4.png \
		warp-door-orange.doodad
	doodad edit-doodad -q --tag color=orange warp-door-orange.doodad
	doodad install-script warp-door.js warp-door-orange.doodad

	for i in *.doodad; do\
		doodad edit-doodad --tag "category=doors" --hitbox=34,76 $${i};\
	done
	for i in warp-door-*.doodad; do\
		doodad edit-doodad --tag "category=doors,gizmos" $${i};\
	done

	cp *.doodad ../../../assets/doodads/
