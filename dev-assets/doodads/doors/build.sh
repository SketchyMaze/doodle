# doodad convert -t "Red Door" red1.png red2.png door-red.doodad
# doodad edit-doodad -q --tag color=red door-red.doodad
# doodad install-script locked-door.js door-red.doodad

doodad convert -t "Red Door" red-closed.png red-unlocked.png red-right.png red-left.png door-red.doodad
doodad edit-doodad -q --tag color=red door-red.doodad
doodad install-script colored-door.js door-red.doodad

doodad convert -t "Blue Door" blue-closed.png blue-unlocked.png blue-right.png blue-left.png door-blue.doodad
doodad edit-doodad -q --tag color=blue door-blue.doodad
doodad install-script colored-door.js door-blue.doodad

doodad convert -t "Green Door" green-closed.png green-unlocked.png green-right.png green-left.png door-green.doodad
doodad edit-doodad -q --tag color=green door-green.doodad
doodad install-script colored-door.js door-green.doodad

doodad convert -t "Yellow Door" yellow-closed.png yellow-unlocked.png yellow-right.png yellow-left.png door-yellow.doodad
doodad edit-doodad -q --tag color=yellow door-yellow.doodad
doodad install-script colored-door.js door-yellow.doodad

doodad convert -t "Small Key Door" small-closed.png small-unlocked.png small-right.png small-left.png small-key-door.doodad
doodad edit-doodad -q --tag color=small small-key-door.doodad
doodad install-script colored-door.js small-key-door.doodad

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

doodad convert -t "Small Key" small-key.png small-key.doodad
doodad edit-doodad -q --tag color=small small-key.doodad
doodad install-script keys.js small-key.doodad

doodad convert -t "Electric Door" electric{1,2,3,4}.png door-electric.doodad
doodad install-script electric-door.js door-electric.doodad

# Tag the category for these doodads
for i in *.doodad; do doodad edit-doodad --tag "category=doors" $i; done
doodad edit-doodad --tag "category=doors,gizmos" door-electric.doodad

cp *.doodad ../../../assets/doodads/
