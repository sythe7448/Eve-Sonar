package eveSolarSystems

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strings"
)

type SolarSystem struct {
	ID          string
	Name        string
	Coordinates Coordinates
	Sec         float64
}

type Coordinates struct {
	X, Y, Z float64
}

const (
	capitalLightYears      float64 = 66225113308060300
	superCapitalLightYears float64 = 56764382835480260
	industryLightYears     float64 = 94607304725800420
	blopsLightYears        float64 = 75685843780640350
)

func GetStagingSystemsBySelectedRangeText(shipRangesSettings ShipRangeSettings, currentSolarSystem SolarSystem) string {
	shipRangesMap := map[string]float64{
		"Blops":    blopsLightYears,
		"Supers":   superCapitalLightYears,
		"Capitals": capitalLightYears,
		"Industry": industryLightYears,
	}

	returnText := ""
	shipRanges := reflect.ValueOf(shipRangesSettings)
	for i := 0; i < shipRanges.NumField(); i++ {
		field := shipRanges.Field(i)
		if field.Bool() {
			fieldString := fmt.Sprintf("%s", shipRanges.Type().Field(i).Name)
			stagingsInRange := QueryStagingsInRange(currentSolarSystem.Coordinates, shipRangesMap[fieldString])
			returnText += fmt.Sprintf("Staging Systems in %s range:\n", fieldString)
			for s, o := range stagingsInRange {
				if s == "" {
					returnText += fmt.Sprintf("No Staging System are in range of %s", fieldString)
				}
				returnText += fmt.Sprintf("%s: %s\n", s, o)
			}
			returnText += "\n"
		}
	}

	return returnText
}

func ConvertStagingSystemsToSting() string {
	systemsString := ""
	stagingSystemsMap := getStagingSystems()
	if len(stagingSystemsMap) != 0 {
		for system, owner := range stagingSystemsMap {
			systemsString += fmt.Sprintf("%s:%s\n", system, owner)
		}
	}
	return systemsString
}

func ParseAndSaveStagingSystems(stagingSystemsText string) {
	if stagingSystemsText == "" {
		err := UpdateStagingSystems(nil)
		if err != nil {
			return
		}
	}
	lines := strings.Split(stagingSystemsText, "\n")
	stagingSystemsMap := make(map[string]string)
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			// Make sure system exists to be added
			exists := QueryForSystemByName(parts[0]).Name
			if len(exists) > 0 {
				stagingSystemsMap[parts[0]] = parts[1]
			}
		}
	}
	err := UpdateStagingSystems(stagingSystemsMap)
	if err != nil {
		return
	}
}

// distance3D calculate the distance in 3d space between 2 points
func Distance3D(p1, p2 Coordinates) float64 {
	dx := bigMathSub(p1.X, p2.X)
	dy := bigMathSub(p1.Y, p2.Y)
	dz := bigMathSub(p1.Z, p2.Z)

	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// bigMathSub this helper is required as the floats are too big to do just math on
func bigMathSub(x float64, y float64) float64 {
	xBig := new(big.Float).SetPrec(256).SetFloat64(x)
	yBig := new(big.Float).SetPrec(256).SetFloat64(y)
	result := new(big.Float).Sub(xBig, yBig)
	float64Result, _ := result.Float64()

	return float64Result
}
