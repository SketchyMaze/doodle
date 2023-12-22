module git.kirsle.net/SketchyMaze/doodle

go 1.16

require (
	git.kirsle.net/go/audio v0.0.0-20230310065553-fa6eb3d3a029
	git.kirsle.net/go/log v0.0.0-20200902035305-70ac2848949b
	git.kirsle.net/go/render v0.0.0-20220505053906-129a24300dfa
	git.kirsle.net/go/ui v0.0.0-20231209035443-e912e2bd035c
	github.com/aichaos/rivescript-go v0.4.0
	github.com/cpuguy83/go-md2man/v2 v2.0.3 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dlclark/regexp2 v1.10.0 // indirect
	github.com/dop251/goja v0.0.0-20231027120936-b396bb4c349d
	github.com/fsnotify/fsnotify v1.7.0
	github.com/gen2brain/dlgs v0.0.0-20220603100644-40c77870fa8d
	github.com/google/pprof v0.0.0-20231205033806-a5a03c77bf08 // indirect
	github.com/google/uuid v1.4.0
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/kirsle/configdir v0.0.0-20170128060238-e45d2f54772f
	github.com/tomnomnom/xtermcolor v0.0.0-20160428124646-b78803f00a7e // indirect
	github.com/urfave/cli/v2 v2.26.0
	github.com/veandco/go-sdl2 v0.4.36
	golang.org/x/crypto v0.16.0 // indirect
	golang.org/x/image v0.14.0
)

replace git.kirsle.net/go/render => /home/kirsle/SketchyMaze/doodle/deps/render

replace git.kirsle.net/go/ui => /home/kirsle/SketchyMaze/doodle/deps/ui

replace git.kirsle.net/go/audio => /home/kirsle/SketchyMaze/doodle/deps/audio
