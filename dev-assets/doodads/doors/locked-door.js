function main() {
	Self.AddAnimation("open", 0, [1]);
	var unlocked = false;

	Events.OnCollide(function(e) {
		console.log("%s was touched by %s!", Self.Doodad.Title, e.Actor.ID());
		console.log("my box: %+v and theirs: %+v", Self.GetBoundingRect(), e.Actor.GetBoundingRect());
		console.warn("But the overlap is: %+v", e.Overlap);
		console.log(Object.keys(e));

		if (e.Overlap.X + e.Overlap.W >= 16 && e.Overlap.X < 48) {
			Self.Canvas.SetBackground(RGBA(255, 0, 0, 153));
		} else {
			Self.Canvas.SetBackground(RGBA(0, 255, 0, 153));
			return;
		}

		if (unlocked) {
			return;
		}

		unlocked = true;
		Self.PlayAnimation("open", null);
	});
	Events.OnLeave(function(e) {
		console.log("%s has stopped touching %s", e, Self.Doodad.Title)
		Self.Canvas.SetBackground(RGBA(0, 0, 1, 0));
	})
}
