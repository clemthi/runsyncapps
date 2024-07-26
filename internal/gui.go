package internal

import (
	"github.com/getlantern/systray"
)

func OnReadyUI() {
	systray.SetTemplateIcon(SysTrayIcon, SysTrayIcon)
	systray.SetTitle("RunSyncApps")
	systray.SetTooltip("RunSyncApps")

	// Exit menu
	mQuitOrig := systray.AddMenuItem("Quit", "Quit")
	go func() {
		<-mQuitOrig.ClickedCh
		systray.Quit()
	}()
}
