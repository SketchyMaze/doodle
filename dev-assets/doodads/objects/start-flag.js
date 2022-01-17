// Start Flag.
function main() {
    Self.SetHitbox(22 + 16, 16, 75 - 16, 86);

    // Linking a doodad to the Start Flag sets the
    // player character. Destroy the original doodads.
    for (var link of Self.GetLinks()) {
        link.Destroy();
    }
}
