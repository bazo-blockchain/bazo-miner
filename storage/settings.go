package storage

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

type Settings struct {
	SeedFileName string `json:"seed-file-name"`
	KeyFileName  string `json:"key-file-name"`
}

func GetSettings(fileName string) (Settings, error) {
	var settings Settings

	//check if file exists
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return settings, errors.New("Settings file does not exist")
	}

	raw, err := ioutil.ReadFile(fileName)
	if err != nil {
		return settings, err
	}

	json.Unmarshal(raw, &settings)

	return settings, nil
}

func WriteSettings(settings Settings, fileName string) error {
	b, err := json.Marshal(settings)

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
