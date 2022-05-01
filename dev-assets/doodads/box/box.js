// Pushable Box.

const speed = 4,
    size = 75;

function main() {
    Self.SetMobile(true);
    Self.SetInvulnerable(true);
    Self.SetGravity(true);
    Self.SetHitbox(0, 0, size, size);

    Events.OnCollide((e) => {
        // Ignore events from neighboring Boxes.
        if (e.Actor.Actor.Filename === "box.doodad") {
            return false;
        }

        if (e.Actor.IsMobile() && e.InHitbox) {
            let overlap = e.Overlap;

            if (overlap.Y === 0 && !(overlap.X === 0 && overlap.W < 5) && !(overlap.X === size)) {
                // Be sure to position them snug on top.
                // TODO: this might be a nice general solution in the
                // collision detector...
                console.log("new box code");
                e.Actor.MoveTo(Point(
                    e.Actor.Position().X,
                    Self.Position().Y - e.Actor.Hitbox().Y - e.Actor.Hitbox().H - 2,
                ))
                e.Actor.SetGrounded(true);
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
    let origPoint = Self.Position();
    Message.Subscribe("power", (powered) => {
        Self.MoveTo(origPoint);
        Self.SetVelocity(Vector(0, 0));
    });

    // Start animation on a loop.
    animate();
}

function animate() {
    Self.AddAnimation("animate", 100, [0, 1, 2, 3, 2, 1]);

    let running = false;
    setInterval(() => {
        if (!running) {
            running = true;
            Self.PlayAnimation("animate", function () {
                running = false;
            })
        }
    }, 100);
}
