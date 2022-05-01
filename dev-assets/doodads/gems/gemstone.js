// Gem stone collectibles/keys.

const color = Self.GetTag("color"),
    shimmerFreq = 1000;

function main() {
	Self.SetMobile(true);
	Self.SetGravity(true);

    Self.AddAnimation("shimmer", 100, [0, 1, 2, 3, 0]);
	Events.OnCollide((e) => {
		if (e.Settled) {
			if (e.Actor.HasInventory()) {
				Sound.Play("item-get.wav")
				e.Actor.AddItem(Self.Filename, 1);
				Self.Destroy();
			}
		}
	});

    setInterval(() => {
        Self.PlayAnimation("shimmer", null);
    }, shimmerFreq);
}