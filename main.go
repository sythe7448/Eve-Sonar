package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"stagingRangeWarning/eveSolarSystems"
)

func main() {
	trackerApp := app.New()
	trackerApp.Settings().SetTheme(theme.DarkTheme())
	trackerWindow := trackerApp.NewWindow("Eve Staging Range Tracker")

	appContainer := eveSolarSystems.BuildContainer(trackerApp)

	trackerWindow.SetContent(appContainer)
	trackerWindow.ShowAndRun()
}
