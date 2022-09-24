package modal_test

import (
	"fmt"

	modal "git.kirsle.net/SketchyMaze/doodle/pkg/modal"
)

func ExampleAlert() {
	alert := modal.Alert("Permission Denied").WithTitle("Error").Then(func() {
		fmt.Println("Alert button answered!")
	})

	_ = alert
}
