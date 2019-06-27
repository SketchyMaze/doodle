# WebAssembly Port

## Build and Test

Change to the wasm/ folder and type `make` to build `doodle.wasm`

To test it with a local Go server, cd to the wasm/ folder and run
`go run server.go` and visit http://localhost:8080/

Copy the `fonts` and `assets` folders from the project root to the
wasm/ directory so they're available over HTTP.

## wasm_exec.js

To update the wasm_exec.js to match your version of Go:

```bash
# Fedora: install golang-misc
sudo dnf install golang-misc

# Copy the wasm_exec.js
cp $(go env GOROOT)/misc/wasm/wasm_exec.js ./
```

## Font Support

Fonts are implemented as CSS embedded fonts configured in
`wasm/index.html`

The font family name should match the filename, sans .ttf extension,
for example "DejaVuSans-Bold". Doodle's internal logic converts a
FontFilename string like "./fonts/DejaVuSans.ttf" into the base name
to use as the font family. It also has fallbacks for sans-serif and
serif in case of any problems.

## Known Bugs and Limitations

* github.com/kirsle/golog
  * The detection of an interactive terminal is broken in WASM.
  * `terminal.IsTerminal(int(os.Stdout.Fd()))`
  * As a workaround, comment it out and hardcode to `false`
* Userdir
  * For WASM we'll want to use localStorage to store user drawings
    instead of the userdir.
* Wallpaper support
  * WASM can't use os.Open to read the wallpaper image, so will need
    another method to load the image.
  * Texture caching support isn't implemented yet to hold the four
    corner textures of a wallpaper.
  * As a workaround, added a `wallpaper.ready` boolean and relevant
    functions exit early for WASM so wallpapers don't render at all.
