// Generic Item Script
/*
A script that makes your item pocket-able, like the Keys.

Your doodad sprite will appear in the Inventory menu if the
player picks it up.

Configure it with tags:
- quantity: integer quantity value, default is 1,
            set to 0 to make it a 'key item'

Can be attached to any doodad.
*/

function main() {
    // Make the hitbox be the full canvas size of this doodad.
    // Adjust if you want a narrower hitbox.
    if (Self.Hitbox().IsZero()) {
        var size = Self.Size()
        Self.SetHitbox(0, 0, size.W, size.H)
    }

    var qtySetting = Self.GetTag("quantity")
    var quantity = qtySetting === "" ? 1 : parseInt(qtySetting);

    Events.OnCollide(function (e) {
        if (e.Settled) {
            if (e.Actor.HasInventory()) {
                Sound.Play("item-get.wav")
                e.Actor.AddItem(Self.Filename, quantity);
                Self.Destroy();
            }
        }
    })
}
