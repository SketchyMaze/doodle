ALL: build

.PHONY: build
build:
	# Goal Region
	doodad convert -t "Goal Region" goal-128.png reg-goal.doodad
	doodad install-script goal.js reg-goal.doodad

	# Checkpoint Region
	doodad convert -t "Checkpoint Region" checkpoint-128.png reg-checkpoint.doodad
	doodad install-script checkpoint.js reg-checkpoint.doodad

	# Fire Region
	doodad convert -t "Fire Region" fire-128.png reg-fire.doodad
	doodad install-script fire.js reg-fire.doodad

	# Stall Region
	doodad convert -t "Stall Player (250ms)" stall-128.png reg-stall-250.doodad
	doodad edit-doodad --tag "ms=250" reg-stall-250.doodad
	doodad install-script stall.js reg-stall-250.doodad

	# Power Source
	doodad convert -t "Power Source" power-64.png power-source.doodad
	doodad install-script power.js power-source.doodad

	# Warp Door
	doodad convert -t "Invisible Warp Door" warp-door-64.png reg-warp-door.doodad
	doodad edit-doodad --tag "color=invisible" reg-warp-door.doodad
	doodad install-script ../warp-door/warp-door.js reg-warp-door.doodad

	# Reset Level Timer
	doodad convert -t "Reset Level Timer" timer-64.png reg-reset-timer.doodad
	doodad install-script reset-timer.js reg-reset-timer.doodad

	for i in *.doodad; do\
		doodad edit-doodad --tag "category=technical" $${i};\
	done
	cp *.doodad ../../../assets/doodads/
