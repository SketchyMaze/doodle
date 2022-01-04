module git.kirsle.net/apps/doodle

go 1.16

require (
	git.kirsle.net/go/audio v0.0.0-20200429055451-ae3b0695ba6f
	git.kirsle.net/go/log v0.0.0-20200902035305-70ac2848949b
	git.kirsle.net/go/render v0.0.0-20211231003948-9e640ab5c3da
	git.kirsle.net/go/ui v0.0.0-20200710023146-e2a561fbd04c
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gen2brain/dlgs v0.0.0-20210406143744-f512297a108e
	github.com/google/uuid v1.1.2
	github.com/gopherjs/gopherjs v0.0.0-20210603182125-eeedf4a0e899 // indirect
	github.com/kirsle/configdir v0.0.0-20170128060238-e45d2f54772f
	github.com/robertkrimen/otto v0.0.0-20200922221731-ef014fd054ac
	github.com/tomnomnom/xtermcolor v0.0.0-20160428124646-b78803f00a7e // indirect
	github.com/urfave/cli/v2 v2.3.0
	github.com/veandco/go-sdl2 v0.4.10
	github.com/vmihailenco/msgpack v3.3.3+incompatible
	golang.org/x/image v0.0.0-20211028202545-6944b10bf410
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
)

replace git.kirsle.net/go/render => /home/kirsle/SketchyMaze/deps/render

replace git.kirsle.net/go/ui => /home/kirsle/SketchyMaze/deps/ui

replace git.kirsle.net/go/audio => /home/kirsle/SketchyMaze/deps/audio

//replace git.kirsle.net/go/render => /run/build/sketchymaze/deps/render
//replace git.kirsle.net/go/ui => /run/build/sketchymaze/deps/ui
//replace git.kirsle.net/go/audio => /run/build/sketchymaze/deps/audio
