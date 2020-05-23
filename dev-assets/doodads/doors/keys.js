function main() {
	var color = Self.GetTag("color");

	Events.OnCollide(function(e) {
		if (e.Settled) {
			Sound.Play("item-get.wav")
			e.Actor.AddItem(Self.Filename, 0);
			Self.Destroy();
		}
	})
}
