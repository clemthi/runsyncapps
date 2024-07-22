package internal

import (
	"github.com/getlantern/systray"
)

func OnReadyUI() {
	systray.SetTemplateIcon(SysTrayIcon, SysTrayIcon)
	systray.SetTitle("RunSyncApp")
	systray.SetTooltip("RunSyncApp")

	// Exit menu
	mQuitOrig := systray.AddMenuItem("Quit", "Quit")
	go func() {
		<-mQuitOrig.ClickedCh
		systray.Quit()
	}()
}
