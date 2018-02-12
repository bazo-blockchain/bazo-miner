package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

type SeedJson struct {
	HashedSeed string `json:"hashed-seed"`
	Seed       string `json:"seed"`
}

func GetSeeds(fileName string) ([]SeedJson, error) {
	//create file if not existent
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		_, err := os.Create(fileName)
		if err != nil {
			return nil, err
		}
	}

	raw, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var seedJsons []SeedJson
	json.Unmarshal(raw, &seedJsons)

	return seedJsons, err
}

func GetSeed(hashedSeed [32]byte, fileName string) ([32]byte, error) {
	seeds, err := GetSeeds(fileName)
	if err != nil {
		panic("Cannot validate without seed")
		return [32]byte{}, err
	}

	for _, s := range seeds {
		if s.HashedSeed == fmt.Sprintf("%x", hashedSeed) {
			var seedBuff [32]byte
			copy(seedBuff[:], s.Seed)
			return seedBuff, nil
		}
	}

	return [32]byte{}, errors.New("No seed found with the given Hash")
}

func writeJson(seeds []SeedJson, fileName string) error {

	b, err := json.Marshal(seeds)

	if err != nil {
		return err
	}

	//create file if not existent
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		_, err := os.Create(fileName)
		if err != nil {
			return err
		}
	}

	err = ioutil.WriteFile(fileName, b, 0644)

	return nil
}

func AppendNewSeed(fileName string, seed SeedJson) error {
	//create file if not existent
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		_, err := os.Create(fileName)
		if err != nil {
			return err
		}
	}

	//load previous seeds
	seeds, err := GetSeeds(fileName)
	if err != nil {
		return err
	}

	//append new seed
	seeds = append(seeds, seed)

	//overwrite old file
	err = writeJson(seeds, fileName)
	if err != nil {
		return err
	}

	return nil
}
