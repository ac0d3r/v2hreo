package main

import (
	"v2rayss/ui"

	"github.com/getlantern/systray"
)

func main() {
	systray.Run(ui.OnReady, ui.OnExit)
}
