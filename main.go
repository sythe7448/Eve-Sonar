package main

import (
	"encoding/csv"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type SolarSystem struct {
	Name string
	Coordinates
	Sec float64
}

type Coordinates struct {
	X, Y, Z *big.Float
}

func main() {
	//variables to use later
	capitalLightYears, _ := new(big.Float).SetString("6.6225113308060300e16")
	superCapitalLightYears, _ := new(big.Float).SetString("5.6764382835480260e16")
	industryLightyears, _ := new(big.Float).SetString("9.4607304725800420e16")
	blopsLightYears, _ := new(big.Float).SetString("7.5685843780640350e16")
	blops := true
	supers := true
	capitals := true
	rorqs := true

	solarSystems := GetEveSolarSystems()

	currentSolarSystem := solarSystems["Turnur"]
	if currentSolarSystem.Sec > .45 {
		fmt.Println("You're in highsec nothing can cyno to you.")
	} else {
		systemsInRange := make(map[string]string)
		if blops {
			systemsInRange = getSystemsInRange(solarSystems, currentSolarSystem.Coordinates, blopsLightYears)
			blopsStagingsInRange := getStagingsInRange(systemsInRange)
			fmt.Println("Staging Systems in blops range:")
			for s, o := range blopsStagingsInRange {
				fmt.Println(s + ":" + o)
			}
		}
		if supers {
			systemsInRange = getSystemsInRange(solarSystems, currentSolarSystem.Coordinates, superCapitalLightYears)
			supersStagingsInRange := getStagingsInRange(systemsInRange)
			fmt.Println("Staging Systems in super range:")
			for s, o := range supersStagingsInRange {
				fmt.Println(s + ":" + o)
			}
		}
		if capitals {
			systemsInRange = getSystemsInRange(solarSystems, currentSolarSystem.Coordinates, capitalLightYears)
			capitalsStagingsInRange := getStagingsInRange(systemsInRange)
			fmt.Println("Staging Systems in capital range:")
			for s, o := range capitalsStagingsInRange {
				fmt.Println(s + ":" + o)
			}
		}
		if rorqs {
			systemsInRange = getSystemsInRange(solarSystems, currentSolarSystem.Coordinates, industryLightyears)
			rorqsStagingsInRange := getStagingsInRange(systemsInRange)
			fmt.Println("Staging Systems in rorqual range:")
			for s, o := range rorqsStagingsInRange {
				fmt.Println(s + ":" + o)
			}
		}
	}

}

func getStagingsInRange(systemsInRange map[string]string) map[string]string {
	// Temp harded coded inputs
	stagingSystems := make(map[string]string)
	stagingSystems["Amamake"] = "Pandemic Legion"
	stagingSystems["Jita"] = "Pubbies"
	stagingSystems["Kurniainen"] = "Amarr Militia"

	stagingInRange := make(map[string]string)
	for system, owner := range stagingSystems {
		if _, exists := systemsInRange[strings.ToLower(system)]; exists {
			stagingInRange[system] = owner
		}
	}

	return stagingInRange

}

func getSystemsInRange(solarSystems map[string]SolarSystem, currentSystemData Coordinates, jumpRange *big.Float) map[string]string {
	radiusSquared := new(big.Float).Mul(jumpRange, jumpRange)

	systemsInRange := make(map[string]string)

	for _, solarSystem := range solarSystems {
		if solarSystem.Coordinates != currentSystemData && squaredDistance3D(currentSystemData, solarSystem.Coordinates).Cmp(radiusSquared) <= 0 {
			systemsInRange[strings.ToLower(solarSystem.Name)] = strings.ToLower(solarSystem.Name)
		}
	}

	return systemsInRange

}

func squaredDistance3D(p1, p2 Coordinates) *big.Float {
	dx := new(big.Float).Sub(p1.X, p2.X)
	dy := new(big.Float).Sub(p1.Y, p2.Y)
	dz := new(big.Float).Sub(p1.Z, p2.Z)

	dxSquared := new(big.Float).Mul(dx, dx)
	dySquared := new(big.Float).Mul(dy, dy)
	dzSquared := new(big.Float).Mul(dz, dz)

	return new(big.Float).Add(dxSquared, new(big.Float).Add(dySquared, dzSquared))
}

func GetEveSolarSystems() map[string]SolarSystem {
	solarSystemsFile, err := os.OpenFile("eveSolarSystems.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
	defer solarSystemsFile.Close()
	csvReader := csv.NewReader(solarSystemsFile)
	if _, err := csvReader.Read(); err != nil {
		panic(err)
	}

	csvData, err := csvReader.ReadAll()
	if err != nil {
		panic(err)
	}

	// Define a regular expression pattern
	pattern := `J[0-9]{6}`

	// Compile the regular expression
	regex, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Println("Error compiling regex:", err)
	}

	// format data for fast access
	solarSystems := make(map[string]SolarSystem)
	for _, data := range csvData {
		// remove WHs
		if regex.MatchString(data[0]) {

		} else {
			coords := make(map[int]*big.Float)
			for i := 1; i < 4; i++ {
				coords[i], _ = new(big.Float).SetString(data[i])
			}
			sec, err := strconv.ParseFloat(data[4], 64)
			if err != nil {
				fmt.Println("Error:", err)
			}
			solarSystems[data[0]] = SolarSystem{Name: data[0], Sec: sec, Coordinates: Coordinates{X: coords[1], Y: coords[2], Z: coords[3]}}
		}
	}

	return solarSystems
}
