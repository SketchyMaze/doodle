function main() {
	var color = Self.GetTag("color");
	var quantity = color === "small" ? 1 : 0;

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
