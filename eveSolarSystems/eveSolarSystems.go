package eveSolarSystems

import (
	"encoding/csv"
	"fmt"
	"math"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type SolarSystem struct {
	ID          int64
	Name        string
	Coordinates Coordinates
	Sec         float64
}

type Coordinates struct {
	X, Y, Z float64
}

type ShipRangeSettings struct {
	Blops, Supers, Capitals, Industry bool
}

var SolarSystemsByNameMap = make(map[string]SolarSystem)
var SolarSystemsByIdMap = make(map[int64]SolarSystem)

const (
	capitalLightYears      float64 = 66225113308060300
	superCapitalLightYears float64 = 56764382835480260
	industryLightYears     float64 = 94607304725800420
	blopsLightYears        float64 = 75685843780640350
)

func init() {
	SetEveSolarSystems()
}

func PrintStagingSystemsBySelectedRange(shipRanges ShipRangeSettings, currentSolarSystem SolarSystem) {
	systemsInRange := make(map[string]struct{})
	if shipRanges.Blops {
		systemsInRange = getSystemsInRange(SolarSystemsByNameMap, currentSolarSystem.Coordinates, blopsLightYears)
		blopsStagingsInRange := getStagingsInRange(systemsInRange)
		fmt.Println("Staging Systems in blops range:")
		for s, o := range blopsStagingsInRange {
			if s == "" {
				fmt.Printf("No Staging System are in range of blops")
			}
			fmt.Printf("%s:%s\n", s, o)
		}
	}
	if shipRanges.Supers {
		systemsInRange = getSystemsInRange(SolarSystemsByNameMap, currentSolarSystem.Coordinates, superCapitalLightYears)
		supersStagingsInRange := getStagingsInRange(systemsInRange)
		fmt.Println("Staging Systems in super range:")
		for s, o := range supersStagingsInRange {
			if s == "" {
				fmt.Printf("No Staging System are in range of Supers")
			}
			fmt.Printf("%s:%s\n", s, o)
		}
	}
	if shipRanges.Capitals {
		systemsInRange = getSystemsInRange(SolarSystemsByNameMap, currentSolarSystem.Coordinates, capitalLightYears)
		capitalsStagingsInRange := getStagingsInRange(systemsInRange)
		fmt.Println("Staging Systems in capital range:")
		for s, o := range capitalsStagingsInRange {
			if s == "" {
				fmt.Printf("No Staging System are in range of Capitals")
			}
			fmt.Printf("%s:%s\n", s, o)
		}
	}
	if shipRanges.Industry {
		systemsInRange = getSystemsInRange(SolarSystemsByNameMap, currentSolarSystem.Coordinates, industryLightYears)
		rorqsStagingsInRange := getStagingsInRange(systemsInRange)
		fmt.Println("Staging Systems in rorqual range:")
		for s, o := range rorqsStagingsInRange {
			if s == "" {
				fmt.Printf("No Staging System are in range of Rorqs")
			}
			fmt.Printf("%s:%s\n", s, o)
		}
	}
}

// getStagingsInRange see if the user inputted solar system is in the map of systems in range
func getStagingsInRange(systemsInRange map[string]struct{}) map[string]string {
	stagingSystems := getStagingSystems()
	stagingInRange := make(map[string]string)
	for system, owner := range stagingSystems {
		if _, exists := systemsInRange[strings.ToLower(system)]; exists {
			stagingInRange[system] = owner
		}
	}

	return stagingInRange
}

// getStagingSystems Get the user inputted staging solar system names
func getStagingSystems() map[string]string {
	// Temp harded coded inputs
	stagingSystems := make(map[string]string)
	stagingSystems["Amamake"] = "Pandemic Legion"
	stagingSystems["Jita"] = "Pubbies"
	stagingSystems["Kurniainen"] = "Amarr Militia"
	stagingSystems["Poitot"] = "The only named system in syndicate"

	return stagingSystems
}

// getSystemsInRange make a map of the solar system names with in a radius to another solar system
func getSystemsInRange(solarSystems map[string]SolarSystem, currentSystemData Coordinates, jumpRange float64) map[string]struct{} {
	systemsInRange := make(map[string]struct{})
	for _, solarSystem := range solarSystems {
		if solarSystem.Coordinates != currentSystemData && distance3D(currentSystemData, solarSystem.Coordinates) <= jumpRange {
			systemsInRange[strings.ToLower(solarSystem.Name)] = struct{}{}
		}
	}

	return systemsInRange
}

func GetSolarSystemByName(systemName string) SolarSystem {
	return SolarSystemsByNameMap[strings.ToLower(systemName)]
}

func GetSolarSystemById(systemId int64) SolarSystem {
	return SolarSystemsByIdMap[systemId]
}

// distance3D calculate the distance in 3d space between 2 points
func distance3D(p1, p2 Coordinates) float64 {
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

// SetEveSolarSystems Opens the hardcoded CSV to create a struct of the solar system data
func SetEveSolarSystems() {
	solarSystemsFile, err := os.OpenFile("eveSolarSystems/eveSolarSystems.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("Error opening the solar system CSV: %s\n", err))
	}
	defer solarSystemsFile.Close()
	csvReader := csv.NewReader(solarSystemsFile)
	// Skips the headers of the CSV
	if _, err := csvReader.Read(); err != nil {
		panic(fmt.Sprintf("Error skipping the first row in the CSV: %s\n", err))
	}

	// Reads the rest of the CSV
	csvData, err := csvReader.ReadAll()
	if err != nil {
		panic(fmt.Sprintf("Error Reading the solar system CSV: %s\n", err))
	}

	// Compile the regular expression
	regex, err := regexp.Compile(`J[0-9]{6}`)
	if err != nil {
		fmt.Println("Error compiling regex:", err)
	}

	// format data for fast access
	for _, data := range csvData {
		// remove WHs
		if regex.MatchString(data[1]) {
			continue
		}
		coords := make(map[int]float64)
		for i := 2; i < 5; i++ {
			coords[i], err = strconv.ParseFloat(data[i], 64)
			if err != nil {
				panic(fmt.Sprintf("Error Parsing %s coordinate float: %s\n", data[0], err))
			}
		}
		sec, err := strconv.ParseFloat(data[5], 64)
		if err != nil {
			panic(fmt.Sprintf("Error Parsing %s sec float:  %s\n", data[0], err))
		}
		id, err := strconv.ParseInt(data[0], 10, 64)
		if err != nil {
			panic(fmt.Sprintf("Error Parsing %s id float:  %s\n", data[0], err))
		}
		SolarSystemsByNameMap[strings.ToLower(data[1])] = SolarSystem{
			ID:   id,
			Name: data[1],
			Sec:  sec,
			Coordinates: Coordinates{
				X: coords[2],
				Y: coords[3],
				Z: coords[4],
			},
		}
		SolarSystemsByIdMap[id] = SolarSystem{
			ID:   id,
			Name: data[1],
			Sec:  sec,
			Coordinates: Coordinates{
				X: coords[2],
				Y: coords[3],
				Z: coords[4],
			},
		}
	}
}
