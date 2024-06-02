package main

import (
	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
)

func onReadyUI() {
	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTitle("RunSyncApp")
	systray.SetTooltip("RunSyncApp")

	// Exit menu
	mQuitOrig := systray.AddMenuItem("Quit", "Quit")
	go func() {
		<-mQuitOrig.ClickedCh
		systray.Quit()
	}()
}
