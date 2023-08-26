package main

import (
	"fyne.io/fyne/v2/app"
	"stagingRangeWarning/eveSolarSystems"
)

func main() {
	trackerApp := app.New()
	trackerWindow := trackerApp.NewWindow("Eve Staging Range Tracker")

	appContainer := eveSolarSystems.BuildContainer(trackerApp)

	trackerWindow.SetContent(appContainer)
	trackerWindow.ShowAndRun()
}
