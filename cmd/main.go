package main

import (
	"v2rayss/logs"
	"v2rayss/ui"

	"github.com/getlantern/systray"
)

func main() {
	defer logs.Close()
	systray.Run(ui.OnReady, ui.OnExit)
}
