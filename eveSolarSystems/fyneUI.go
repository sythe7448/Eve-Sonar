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

type ShipRangeSettings struct {
	Blops, Supers, Capitals, Industry bool
}

var rangeSettings = ShipRangeSettings{}
var currentSolarSystemID string
var currentSystemText = widget.NewLabel("")
var stagingInRangeText = widget.NewLabel("")

func BuildContainer(app fyne.App) *fyne.Container {
	// Variables that are passed
	currentSolarSystemID, _ = ESI.GetLocationId(ESI.Tokens.AccessToken, ESI.Character.CharacterID)
	updateCurrentSystemName(currentSystemText, currentSolarSystemID)

	// Set each box
	rangeSettingsBox := buildRangeSettingBox(app)
	stagerSettingBox := buildStagerSettingsBox()
	systemDataBox := container.NewVBox(
		currentSystemText,
		stagingInRangeText,
	)

	// Start a loop to update ranges every 10 seconds
	go func() {
		oldCurrentSolarSystemID := currentSolarSystemID
		for range time.Tick(time.Second * 10) {
			if len(ESI.Tokens.AccessToken) > 0 {
				currentSolarSystemID, _ = ESI.GetLocationId(ESI.Tokens.AccessToken, ESI.Character.CharacterID)
				if oldCurrentSolarSystemID != currentSolarSystemID {
					updateCurrentSystemName(currentSystemText, currentSolarSystemID)
					updateStagerText(rangeSettings, stagingInRangeText, currentSolarSystemID)
					oldCurrentSolarSystemID = currentSolarSystemID
				}
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

func buildRangeSettingBox(app fyne.App) *fyne.Container {
	//build manual system input box
	systemInput := widget.NewEntry()
	systemInput.SetPlaceHolder("Enter system here")
	manualSystemSubmit := widget.NewButton("Check Ranges", func() {
		currentSolarSystemID = QueryForSystemByName(systemInput.Text).ID
		updateCurrentSystemName(currentSystemText, currentSolarSystemID)
		updateStagerText(rangeSettings, stagingInRangeText, currentSolarSystemID)
	})
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
		widget.NewLabel("Manual System Input"),
		systemInput,
		manualSystemSubmit,
		widget.NewLabel("Range options:"),
		blopsCheckBox,
		superCheckBox,
		capitalCheckBox,
		industryCheckBox,
		widget.NewLabel("Login to track location"),
		loginButton,
		widget.NewButton("Quit", func() {
			app.Quit()
		}),
	)

	return rangeSettingsBox

}

func buildStagerSettingsBox() *fyne.Container {
	stagers := widget.NewMultiLineEntry()

	stagers.SetText(ConvertStagingSystemsToSting())

	stagers.SetPlaceHolder("system:owner")
	stagerContainer := container.NewScroll(stagers)
	stagerContainer.SetMinSize(fyne.NewSize(100, 300))
	saveStagers := widget.NewButton("Submit", func() {
		ParseAndSaveStagingSystems(stagers.Text)
		updateStagerText(rangeSettings, stagingInRangeText, currentSolarSystemID)
	})

	stagerSettingBox := container.NewVBox(
		widget.NewLabel("Staging Systems\n system:owner \n new line for new entry"),
		stagerContainer,
		saveStagers,
	)
	return stagerSettingBox
}

func updateCurrentSystemName(currentSystemText *widget.Label, currentSolarSystemID string) {
	if len(currentSolarSystemID) == 0 {
		return
	}
	currentSolarSystemName := QueryForSystemByID(currentSolarSystemID).Name
	currentSystemText.SetText(fmt.Sprintf("Current System: %s", currentSolarSystemName))
}

func updateStagerText(rangeSettings ShipRangeSettings, rangeText *widget.Label, currentSolarSystemID string) {
	if len(currentSolarSystemID) == 0 {
		return
	}
	currentSolarSystem := QueryForSystemByID(currentSolarSystemID)
	rangeText.SetText(GetStagingSystemsBySelectedRangeText(rangeSettings, currentSolarSystem))
}

func openWebpage(urlStr string, app fyne.App) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	return app.OpenURL(u)
}
