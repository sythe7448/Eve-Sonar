package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"log"
	"net/url"
	"stagingRangeWarning/ESI"
	"stagingRangeWarning/eveSolarSystems"
	"time"
)

func main() {
	trackerApp := app.New()
	trackerWindow := trackerApp.NewWindow("Eve Staging Range Tracker")

	loginButton := widget.NewButton("Login to ESI", func() {
		// URL to open
		esiURL := ESI.LocalBaseURI
		go ESI.StartServer()
		// Open the URL in the default web browser
		err := openWebpage(esiURL, trackerApp)
		if err != nil {
			log.Println("Error opening webpage:", err)
		}
	})

	currentSolarSystemID, _ := ESI.GetLocationId(ESI.Tokens.AccessToken, ESI.Character.CharacterID)
	rangeSettings := eveSolarSystems.ShipRangeSettings{}
	currentSystemText := widget.NewLabel("")
	updateCurrentSystemName(currentSystemText, currentSolarSystemID)
	stagingInRangeText := widget.NewLabel("")

	// Create four checkboxes
	blopsCheckBox := widget.NewCheck("Blops Range", func(checked bool) {
		rangeSettings.Blops = checked
		updateStagerText(rangeSettings, stagingInRangeText, currentSolarSystemID)
	})
	superCheckBox := widget.NewCheck("Super Range", func(checked bool) {
		rangeSettings.Supers = checked
		updateStagerText(rangeSettings, stagingInRangeText, currentSolarSystemID)
	})
	capitalCheckBox := widget.NewCheck("Capital Range", func(checked bool) {
		rangeSettings.Capitals = checked
		updateStagerText(rangeSettings, stagingInRangeText, currentSolarSystemID)
	})
	industryCheckBox := widget.NewCheck("Industry Range", func(checked bool) {
		rangeSettings.Industry = checked
		updateStagerText(rangeSettings, stagingInRangeText, currentSolarSystemID)
	})

	trackerWindow.SetContent(currentSystemText)
	go func() {
		for range time.Tick(time.Second * 10) {
			currentSolarSystemID, _ = ESI.GetLocationId(ESI.Tokens.AccessToken, ESI.Character.CharacterID)
			if isNewSystem(currentSolarSystemID) {
				updateCurrentSystemName(currentSystemText, currentSolarSystemID)
				updateStagerText(rangeSettings, stagingInRangeText, currentSolarSystemID)
			}
		}
	}()

	rangeSettingsBox := container.NewVBox(
		widget.NewLabel("Range options:"),
		blopsCheckBox,
		superCheckBox,
		capitalCheckBox,
		industryCheckBox,
	)

	systemDataWindow := container.NewVBox(
		currentSystemText,
		stagingInRangeText,
	)

	loginWindow := container.NewVBox(
		loginButton,
		widget.NewButton("Quit", func() {
			trackerApp.Quit()
		}),
	)

	hbox := container.NewHBox(
		loginWindow,
		rangeSettingsBox,
		systemDataWindow,
	)

	trackerWindow.SetContent(hbox)

	trackerWindow.ShowAndRun()
}

func isNewSystem(currentSolarSystemID int64) bool {
	var oldCurrentSolarSystemID int64
	if oldCurrentSolarSystemID != currentSolarSystemID {
		currentSolarSystemID = oldCurrentSolarSystemID
		return true
	}
	return false
}

func updateCurrentSystemName(currentSystemText *widget.Label, currentSolarSystemID int64) {
	currentSolarSystemName := eveSolarSystems.GetSolarSystemById(currentSolarSystemID).Name
	currentSystemText.SetText(fmt.Sprintf("Current System: %s", currentSolarSystemName))
}

func updateStagerText(rangeSettings eveSolarSystems.ShipRangeSettings, rangeText *widget.Label, currentSolarSystemID int64) {
	currentSolarSystem := eveSolarSystems.GetSolarSystemById(currentSolarSystemID)
	rangeText.SetText(eveSolarSystems.GetStagingSystemsBySelectedRangeText(rangeSettings, currentSolarSystem))
}

func openWebpage(urlStr string, app fyne.App) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	return app.OpenURL(u)
}
