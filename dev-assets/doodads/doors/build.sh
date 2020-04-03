# doodad convert -t "Red Door" red1.png red2.png door-red.doodad
# doodad edit-doodad -q --tag color=red door-red.doodad
# doodad install-script locked-door.js door-red.doodad

doodad convert -t "Red Door" red-closed.png red-right.png red-left.png door-red.doodad
doodad edit-doodad -q --tag color=red door-red.doodad
doodad install-script colored-door.js door-red.doodad

doodad convert -t "Blue Door" blue-closed.png blue-right.png blue-left.png door-blue.doodad
doodad edit-doodad -q --tag color=blue door-blue.doodad
doodad install-script colored-door.js door-blue.doodad

doodad convert -t "Green Door" green-closed.png green-right.png green-left.png door-green.doodad
doodad edit-doodad -q --tag color=green door-green.doodad
doodad install-script colored-door.js door-green.doodad

doodad convert -t "Yellow Door" yellow-closed.png yellow-right.png yellow-left.png door-yellow.doodad
doodad edit-doodad -q --tag color=yellow door-yellow.doodad
doodad install-script colored-door.js door-yellow.doodad

# doodad convert -t "Green Door" green1.png green2.png door-green.doodad
# doodad edit-doodad -q --tag color=green door-green.doodad
# doodad install-script locked-door.js door-green.doodad
#
# doodad convert -t "Yellow Door" yellow1.png yellow2.png door-yellow.doodad
# doodad edit-doodad -q --tag color=yellow door-yellow.doodad
# doodad install-script locked-door.js door-yellow.doodad

doodad convert -t "Red Key" red-key.png key-red.doodad
doodad edit-doodad -q --tag color=red key-red.doodad
doodad install-script keys.js key-red.doodad

doodad convert -t "Blue Key" blue-key.png key-blue.doodad
doodad edit-doodad -q --tag color=blue key-blue.doodad
doodad install-script keys.js key-blue.doodad

doodad convert -t "Green Key" green-key.png key-green.doodad
doodad edit-doodad -q --tag color=green key-green.doodad
doodad install-script keys.js key-green.doodad

doodad convert -t "Yellow Key" yellow-key.png key-yellow.doodad
doodad edit-doodad -q --tag color=yellow key-yellow.doodad
doodad install-script keys.js key-yellow.doodad

doodad convert -t "Electric Door" electric{1,2,3,4}.png door-electric.doodad
doodad install-script electric-door.js door-electric.doodad

cp door-*.doodad key-*.doodad ../../../assets/doodads/
