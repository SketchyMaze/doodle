// Pushable Box.
var speed = 4;
var size = 75;

function main() {
    Self.SetMobile(true);
    Self.SetGravity(true);
    Self.SetHitbox(0, 0, size, size);

    Events.OnCollide(function (e) {
        // Ignore events from neighboring Boxes.
        if (e.Actor.Actor.Filename === "box.doodad") {
            return false;
        }

        if (e.Actor.IsMobile() && e.InHitbox) {
            var overlap = e.Overlap;

            if (overlap.Y === 0 && !(overlap.X === 0 && overlap.W < 5) && !(overlap.X === size)) {
                // Standing on top, ignore.
                return false;
            } else if (overlap.Y === size) {
                // From the bottom, boop it up.
                Self.SetVelocity(Vector(0, -speed * 2))
            }

            // If touching from the sides, slide away.
            if (overlap.X === 0) {
                Self.SetVelocity(Vector(speed, 0))
            } else if (overlap.X === size) {
                Self.SetVelocity(Vector(-speed, 0))
            }

            return false;
        }
    });
    Events.OnLeave(function (e) {
        Self.SetVelocity(Vector(0, 0));
    });

    // When we receive power, we reset to our original position.
    var origPoint = Self.Position();
    Message.Subscribe("power", function (powered) {
        console.error("Box received power! %+v", powered);
        console.error("MoveTo: %+v", origPoint);
        console.error("Keys: %+v", Object.keys(Self));
        Self.MoveTo(origPoint);
        Self.SetVelocity(Vector(0, 0));
    });

    // Start animation on a loop.
    animate();
}

function animate() {
    Self.AddAnimation("animate", 100, [0, 1, 2, 3, 2, 1]);

    var running = false;
    setInterval(function () {
        if (!running) {
            running = true;
            Self.PlayAnimation("animate", function () {
                running = false;
            })
        }
    }, 100);
}