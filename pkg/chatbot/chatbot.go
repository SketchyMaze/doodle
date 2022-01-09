// Package chatbot provides the RiveScript bot that lives in the developer shell.
package chatbot

import (
	"errors"

	"git.kirsle.net/apps/doodle/assets"
	"git.kirsle.net/apps/doodle/pkg/log"
	"github.com/aichaos/rivescript-go"
)

var (
	Bot      *rivescript.RiveScript
	Username = "player"
)

// RiveScript source code snippets to be SURE get included even if
// the external brain fails to load.
const (
	RiveScriptBase = `
		! array boolean = true false t f on off yes no 1 0

		+ echo *
		- <call>echo *</call>

		+ error *
		- <call>error *</call>

		+ alert *
		- <call>alert *</call>

		+ boolprop * (@boolean)
		- <call>boolprop <star1> <star2></call>

		+ rivescript command *
		@ <star>
	`

	RiveScriptBasic = `
		+ hello *
		- Hello <id>.
	`
)

// Bind functions.
type Functions struct {
	Echo        func(string)
	Error       func(string)
	Alert       func(string)
	Confirm     func(string)
	New         func()
	Save        func()
	Edit        func()
	Play        func()
	Close       func()
	TitleScreen func(string)
	BoolProp    func(string, bool)
}

// Setup the RiveScript interpreter.
func Setup() {
	log.Info("Initializing chatbot")
	Bot = rivescript.New(rivescript.WithUTF8())

	if err := Bot.Stream(RiveScriptBase + RiveScriptBasic); err != nil {
		log.Error("Error streaming RiveScript base: %s", err)
	}

	// Load all the built-in .rive scripts.
	if filenames, err := assets.AssetDir("assets/rivescript"); err == nil {
		for _, filename := range filenames {
			if data, err := assets.Asset("assets/rivescript/" + filename); err == nil {
				err = Bot.Stream(string(data))
				if err != nil {
					log.Error("RiveScript.Stream(%s): %s", filename, err)
				}
			} else {
				log.Error("chatbot.Setup: Asset(%s): %s", filename, err)
			}
		}
	} else {
		log.Error("chatbot.Setup: AssetDir: %s", err)
	}

	Bot.SortReplies()
}

// Handle a message from the user.
func Handle(message string) (string, error) {
	if Bot != nil {
		return Bot.Reply(Username, message)
	}
	return "", errors.New("bot not ready")
}
