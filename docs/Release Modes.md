# Release Modes

Verifying your release mode:

```bash
# Full release version
$ doodle -version
doodle version 0.0.7 build 52a2545. Built on 2019-04-19T17:16:19-07:00

# Shareware version includes the word "(shareware)" after the build number.
$ doodle -version
doodle version 0.0.7 build 52a2545 (shareware). Built on 2019-04-19T17:16:19-07:00
```

## Shareware

* `make build-free` to build the shareware binaries to the bin/ folder.

The shareware (free) version of the game has the following restrictions:

* No Doodad Editor Mode available.
  * "New Doodad" button hidden from UI
  * d.NewDoodad() function errors out
  * The dev console `edit <file>` command won't edit a doodad.

## Full Version

The full release of the game has no restrictions and includes the Doodad Editor
Mode.
