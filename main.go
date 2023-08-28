package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"github.com/sythe7448/Eve-Sonar/eveSolarSystems"
)

func main() {
	trackerApp := app.New()
	trackerApp.Settings().SetTheme(theme.DarkTheme())
	trackerWindow := trackerApp.NewWindow("Eve Sonar")

	appContainer := eveSolarSystems.BuildContainer(trackerApp)

	trackerWindow.SetContent(appContainer)
	trackerWindow.ShowAndRun()
}
