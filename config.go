package main

import (
	"encoding/json"
	"os"
)

type AppConfig struct {
	Path                string `json:"path"`
	UseExistingInstance bool   `json:"useExistingInstance"`
	KillOnExit          bool   `json:"killOnExit"`
}

type ConfigFile struct {
	WaitCheck    int         `json:"waitCheck"`
	WaitExit     int         `json:"waitExit"`
	Applications []AppConfig `json:"applications"`
}

func loadConfigFile(configFile string) (*ConfigFile, error) {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var jsonData ConfigFile
	err = json.Unmarshal([]byte(file), &jsonData)

	return &jsonData, err
}
