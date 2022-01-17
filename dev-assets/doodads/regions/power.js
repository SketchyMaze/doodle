// Power source.
// Emits a power(true) signal once on level start.
// If it receives a power signal, it will repeat it after 5 seconds.
// Link two of these bad boys together and you got yourself a clock.
function main() {
    Self.Hide();

    // See if we are not linked to anything.
    var links = Self.GetLinks();
    if (links.length === 0) {
        console.error(
            "%s at %s is not linked to anything! This doodad emits a power(true) on level start to all linked doodads.",
            Self.Title,
            Self.Position()
        );
    }

    Message.Subscribe("broadcast:ready", () => {
        Message.Publish("switch:toggle", true);
        Message.Publish("power", true);
    });
}
