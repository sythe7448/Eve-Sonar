package eveSolarSystems

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/sythe7448/Eve-Sonar/api"
	"log"
	"net/url"
	"strings"
	"time"
)

type ShipRangeSettings struct {
	Blops, Supers, Capitals, Industry bool
}

// variables used locally throughout these functions
var rangeSettings = ShipRangeSettings{}
var currentSolarSystemID string
var currentSystemText = widget.NewLabel("")
var stagingInRangeText = widget.NewLabel("")

// BuildContainer build/design the main container for the app using fyne.
func BuildContainer(app fyne.App) *fyne.Container {
	// Variables that are passed
	currentSolarSystemID, _ = api.GetLocationId(api.Tokens.AccessToken, api.Character.CharacterID)
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
			if len(api.Tokens.AccessToken) > 0 {
				currentSolarSystemID, _ = api.GetLocationId(api.Tokens.AccessToken, api.Character.CharacterID)
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
	// Auto complete
	suggestionList := buildAutoComplete(systemInput)

	manualSystemSubmit := widget.NewButton("Check Ranges", func() {
		currentSolarSystemID = GetSystemByName(systemInput.Text).ID
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
		esiURL := api.LocalBaseURI
		go api.StartServer()
		// Open the URL in the default web browser
		err := openWebpage(esiURL, app)
		if err != nil {
			log.Println("Error opening webpage:", err)
		}
	})

	rangeSettingsBox := container.NewVBox(
		widget.NewLabel("Manual System Input"),
		systemInput,
		suggestionList,
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
	suggestionList := buildAutoComplete(stagers)
	stagerContainer := container.NewScroll(stagers)
	stagerContainer.SetMinSize(fyne.NewSize(100, 350))
	saveStagers := widget.NewButton("Submit", func() {
		ParseAndSaveStagingSystems(stagers.Text)
		updateStagerText(rangeSettings, stagingInRangeText, currentSolarSystemID)
	})

	stagerSettingBox := container.NewVBox(
		widget.NewLabel("Staging Systems\n system:owner \n new line for new entry"),
		suggestionList,
		stagerContainer,
		saveStagers,
	)
	return stagerSettingBox
}

func buildAutoComplete(input *widget.Entry) *fyne.Container {
	suggestionList := container.NewVBox()
	input.OnChanged = func(text string) {
		suggestionList.Objects = nil
		isMultiLine := input.MultiLine
		oldText := ""
		// Filter and populate suggestions based on the user's input
		if isMultiLine {
			lines := strings.Split(text, "\n")
			currentLine := lines[len(lines)-1]
			oldText = func() string {
				if len(lines[:len(lines)-1]) == 0 {
					return ""
				}
				return strings.Join(lines[:len(lines)-1], "\n") + "\n"
			}()
			if len(currentLine) > 2 {
				for _, suggestion := range getSystemSuggestions(currentLine) {
					suggestionItem := widget.NewButton(suggestion, func() {})
					suggestionItem.OnTapped = setText(suggestionItem, input, oldText, suggestionList, isMultiLine)
					suggestionList.Add(suggestionItem)
				}
			}
		} else {
			if len(text) > 2 {
				for _, suggestion := range getSystemSuggestions(text) {
					suggestionItem := widget.NewButton(suggestion, func() {})
					suggestionItem.OnTapped = setText(suggestionItem, input, oldText, suggestionList, isMultiLine)
					suggestionList.Add(suggestionItem)
				}
			}
		}
	}
	return suggestionList
}

func setText(button *widget.Button, input *widget.Entry, oldText string, suggestionList *fyne.Container, isMultiLine bool) func() {
	return func() {
		buttonText := button.Text
		if isMultiLine {
			buttonText += ":"
		}
		input.SetText(oldText + buttonText) // Set selected suggestion with old selects in the input field
		suggestionList.Objects = nil
	}
}

func getSystemSuggestions(prefix string) []string {
	// Sample suggestions for demonstration
	options := GetAllSystems()
	var suggestions []string

	for _, option := range options {
		if strings.HasPrefix(strings.ToLower(option), strings.ToLower(prefix)) {
			suggestions = append(suggestions, option)
		}
	}

	return suggestions
}

func updateCurrentSystemName(currentSystemText *widget.Label, currentSolarSystemID string) {
	if len(currentSolarSystemID) == 0 {
		currentSystemText.SetText(fmt.Sprintf("Current System: No System Found\n If this is a manual input check spelling"))
		return
	}
	currentSolarSystemName := GetSystemByID(currentSolarSystemID).Name
	currentSystemText.SetText(fmt.Sprintf("Current System: %s", currentSolarSystemName))
}

func updateStagerText(rangeSettings ShipRangeSettings, rangeText *widget.Label, currentSolarSystemID string) {
	if len(currentSolarSystemID) == 0 {
		return
	}
	currentSolarSystem := GetSystemByID(currentSolarSystemID)
	rangeText.SetText(GetStagingSystemsBySelectedRangeText(rangeSettings, currentSolarSystem))
}

func openWebpage(urlStr string, app fyne.App) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	return app.OpenURL(u)
}
