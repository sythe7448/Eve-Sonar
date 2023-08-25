package main

import (
	"fmt"
	"stagingRangeWarning/ESI"
	"stagingRangeWarning/eveSolarSystems"
	"time"
)

func main() {
	ESI.StartLogin()

	rangeSettings := eveSolarSystems.ShipRangeSettings{
		Blops:    false,
		Supers:   true,
		Capitals: false,
		Industry: false,
	}

	duration := 5 * time.Minute
	interval := 10 * time.Second
	endTime := time.Now().Add(duration)

	var currentSolarSystemID int64

	for currentTime := time.Now(); currentTime.Before(endTime); currentTime = time.Now() {
		newCurrentSolarSystemID, _ := ESI.GetLocationId(ESI.Tokens.AccessToken, ESI.Character.CharacterID)
		if newCurrentSolarSystemID != currentSolarSystemID {
			currentSolarSystemID = newCurrentSolarSystemID
			currentSolarSystem := eveSolarSystems.GetSolarSystemById(currentSolarSystemID)
			fmt.Printf("You're current system is %s\n", currentSolarSystem.Name)
			if len(currentSolarSystem.Name) == 0 {
				fmt.Println("Your system is not found in known space.")
				continue
			}
			if currentSolarSystem.Sec > .45 {
				fmt.Println("You're in highsec nothing can cyno to you.")
				continue
			} else {
				eveSolarSystems.PrintStagingSystemsBySelectedRange(rangeSettings, currentSolarSystem)
			}
		}
		time.Sleep(interval)
	}
}
