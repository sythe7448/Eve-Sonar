package eveSolarSystems

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"log"
	"net/url"
	"stagingRangeWarning/ESI"
	"time"
)

var rangeSettings = ShipRangeSettings{}

func BuildContainer(app fyne.App) *fyne.Container {
	// Variables that are passed
	currentSolarSystemID, _ := ESI.GetLocationId(ESI.Tokens.AccessToken, ESI.Character.CharacterID)
	currentSystemText := widget.NewLabel("")
	updateCurrentSystemName(currentSystemText, currentSolarSystemID)
	stagingInRangeText := widget.NewLabel("")

	// Set each box
	rangeSettingsBox := buildRangeSettingBox(app, currentSolarSystemID, stagingInRangeText)
	stagerSettingBox := buildStagerSettingsBox()
	systemDataBox := container.NewVBox(
		currentSystemText,
		stagingInRangeText,
	)

	// Start a loop to update ranges every 10 seconds
	go func() {
		for range time.Tick(time.Second * 10) {
			currentSolarSystemID, _ = ESI.GetLocationId(ESI.Tokens.AccessToken, ESI.Character.CharacterID)
			if isNewSystem(currentSolarSystemID) {
				updateCurrentSystemName(currentSystemText, currentSolarSystemID)
				updateStagerText(rangeSettings, stagingInRangeText, currentSolarSystemID)
			}
		}
	}()

	// Build the final lay out to return
	hbox := container.New(
		layout.NewGridLayout(3),
		rangeSettingsBox,
		stagerSettingBox,
		systemDataBox,
	)

	return hbox
}

func buildRangeSettingBox(app fyne.App, currentSolarSystemID int64, stagingInRangeText *widget.Label) *fyne.Container {
	// Build check boxes for ranges
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

	// Login Button
	loginButton := widget.NewButton("Login to ESI", func() {
		// URL to open
		esiURL := ESI.LocalBaseURI
		go ESI.StartServer()
		// Open the URL in the default web browser
		err := openWebpage(esiURL, app)
		if err != nil {
			log.Println("Error opening webpage:", err)
		}
	})

	rangeSettingsBox := container.NewVBox(
		widget.NewLabel("Range options:"),
		blopsCheckBox,
		superCheckBox,
		capitalCheckBox,
		industryCheckBox,
		loginButton,
		widget.NewButton("Quit", func() {
			app.Quit()
		}),
	)

	return rangeSettingsBox

}

func buildStagerSettingsBox() *fyne.Container {
	stagers := widget.NewMultiLineEntry()
	if len(StagingSystemsMap) != 0 {
		stagers.SetText(ConvertStagingSystemsToSting())
	}
	stagers.SetPlaceHolder("system:owner")
	stagerContainer := container.NewScroll(stagers)
	stagerContainer.SetMinSize(fyne.NewSize(100, 300))
	saveStagers := widget.NewButton("Submit", func() {
		ParseAndSaveStagingSystems(stagers.Text)
	})

	stagerSettingBox := container.NewVBox(
		widget.NewLabel("Staging Systems\n system:owner \n new line for new entry"),
		stagerContainer,
		saveStagers,
	)
	return stagerSettingBox
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
	currentSolarSystemName := GetSolarSystemById(currentSolarSystemID).Name
	currentSystemText.SetText(fmt.Sprintf("Current System: %s", currentSolarSystemName))
}

func updateStagerText(rangeSettings ShipRangeSettings, rangeText *widget.Label, currentSolarSystemID int64) {
	currentSolarSystem := GetSolarSystemById(currentSolarSystemID)
	rangeText.SetText(GetStagingSystemsBySelectedRangeText(rangeSettings, currentSolarSystem))
}

func openWebpage(urlStr string, app fyne.App) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	return app.OpenURL(u)
}
