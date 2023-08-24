package main

import (
	"fmt"
	"stagingRangeWarning/eveSolarSystems"
)

func main() {
	rangeSettings := eveSolarSystems.ShipRangeSettings{
		Blops:    true,
		Supers:   false,
		Capitals: false,
		Industry: false,
	}

	currentSolarSystem := eveSolarSystems.GetSolarSystem("Turnur")
	if len(currentSolarSystem.Name) == 0 {
		fmt.Println("Your system is not found in known space.")
		return
	}
	if currentSolarSystem.Sec > .45 {
		fmt.Println("You're in highsec nothing can cyno to you.")
		return
	} else {
		eveSolarSystems.PrintStagingSystemsBySelectedRange(rangeSettings, currentSolarSystem)
	}

}
