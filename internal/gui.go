package internal

import (
	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
)

func OnReadyUI() {
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
