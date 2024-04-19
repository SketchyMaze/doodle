package native

import (
	"os"
)

var (
	USER          string = os.Getenv("USER")
	DefaultAuthor        = USER
)
