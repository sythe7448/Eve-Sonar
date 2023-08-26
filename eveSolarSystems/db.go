package eveSolarSystems

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	DbFile               string = "eveSolarSystems/tracker.db"
	SolarSystemsBucket   string = "solarSystems"
	StagingSystemsBucket string = "stagingSystems"
)

func init() {
	db, err := bolt.Open(DbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// check if bucket exists and build solar system bucket
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(SolarSystemsBucket))
		if bucket == nil {
			bucket, err = tx.CreateBucketIfNotExists([]byte(SolarSystemsBucket))
			if err != nil {
				return err
			}
			solarSystemMap := buildEveSolarSystemsMap()
			err = buildSolarSystemBucket(solarSystemMap, tx, bucket)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		panic(fmt.Sprintf("Error building Solar System DB: %s", err))
	}
	if err != nil {
		log.Fatal(err)
	}
}

func QueryForSystemByID(id string) SolarSystem {
	if id == "" {
		return SolarSystem{}
	}

	db, err := bolt.Open(DbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// convertID to bytes
	idBytes := []byte(id)

	var retrievedSolarSystem SolarSystem

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(SolarSystemsBucket))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		serializedData := bucket.Get(idBytes)
		if serializedData == nil {
			return fmt.Errorf("key not found")
		}

		decoder := gob.NewDecoder(bytes.NewReader(serializedData))
		if err := decoder.Decode(&retrievedSolarSystem); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	if len(retrievedSolarSystem.ID) > 0 {
		return retrievedSolarSystem
	}

	return SolarSystem{}
}

func QueryStagingsInRange(currentSystemData Coordinates, jumpRange float64) map[string]string {
	db, err := bolt.Open(DbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	stagingInRange := make(map[string]string)
	systemsInRange := querySystemsInRange(currentSystemData, jumpRange, db)
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(StagingSystemsBucket))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}
		err = bucket.ForEach(func(system, owner []byte) error {
			if _, exists := systemsInRange[strings.ToLower(string(system))]; exists {
				stagingInRange[string(system)] = string(owner)
			}
			return nil
		})

		if err != nil {
			log.Fatal(err)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	return stagingInRange
}

func querySystemsInRange(currentSystemData Coordinates, jumpRange float64, db *bolt.DB) map[string]struct{} {
	systemsInRange := make(map[string]struct{})
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(SolarSystemsBucket))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}
		err := bucket.ForEach(func(key, value []byte) error {
			var solarSystem SolarSystem
			decoder := gob.NewDecoder(bytes.NewReader(value))
			if err := decoder.Decode(&solarSystem); err != nil {
				return err
			}
			if solarSystem.Coordinates != currentSystemData && Distance3D(currentSystemData, solarSystem.Coordinates) <= jumpRange {
				systemsInRange[strings.ToLower(solarSystem.Name)] = struct{}{}
			}
			return nil
		})

		if err != nil {
			log.Fatal(err)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	return systemsInRange
}

func QueryForSystemByName(name string) SolarSystem {
	var retrievedSolarSystem SolarSystem

	db, err := bolt.Open(DbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(SolarSystemsBucket))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		err := bucket.ForEach(func(key, value []byte) error {
			var solarSystem SolarSystem
			decoder := gob.NewDecoder(bytes.NewReader(value))
			if err := decoder.Decode(&solarSystem); err != nil {
				return err
			}

			if strings.ToLower(solarSystem.Name) == strings.ToLower(name) {
				retrievedSolarSystem = solarSystem
			}
			return nil
		})
		return err
	})

	if err != nil {
		log.Fatal(err)
	}

	if len(retrievedSolarSystem.ID) > 0 {
		return retrievedSolarSystem
	}

	return SolarSystem{}
}

func getStagingSystems() map[string]string {
	db, err := bolt.Open(DbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	stagings := make(map[string]string)
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(StagingSystemsBucket))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}
		err := bucket.ForEach(func(system, owner []byte) error {
			stagings[string(system)] = string(owner)
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	return stagings
}

func UpdateStagingSystems(stagingSystems map[string]string) error {
	db, err := bolt.Open(DbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(StagingSystemsBucket))
		if err != nil {
			return err
		}
		// Iterate through the map and store each struct
		for staging, owner := range stagingSystems {
			// Store the serialized struct in the bucket
			err = bucket.Put([]byte(staging), []byte(owner))
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func buildSolarSystemBucket(solarSystemsByIdMap map[string]SolarSystem, tx *bolt.Tx, bucket *bolt.Bucket) error {
	// Iterate through the map and store each struct
	for id, solarSystem := range solarSystemsByIdMap {
		// Serialize the struct using encoding/gob
		var encodedSolarSystem bytes.Buffer
		enc := gob.NewEncoder(&encodedSolarSystem)
		err := enc.Encode(solarSystem)
		if err != nil {
			return err
		}

		// Store the serialized struct in the bucket
		err = bucket.Put([]byte(id), encodedSolarSystem.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

// SetEveSolarSystems Opens the hardcoded CSV to create a struct of the solar system data
func buildEveSolarSystemsMap() map[string]SolarSystem {
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

	solarSystemsByIdMap := make(map[string]SolarSystem)
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
		solarSystemsByIdMap[data[0]] = SolarSystem{
			ID:   data[0],
			Name: data[1],
			Sec:  sec,
			Coordinates: Coordinates{
				X: coords[2],
				Y: coords[3],
				Z: coords[4],
			},
		}
	}
	return solarSystemsByIdMap
}
