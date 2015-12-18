package main

import (
	"encoding/json"
	"io"
	"os"
)

// ConfigFile config.json file info
type ConfigFile struct {
	Discovery string `json:"discovery"`
	Advertise string `json:"advertise"`
}

func (configFile *ConfigFile) LoadFromReader(configData io.Reader) error {
	if err := json.NewDecoder(configData).Decode(&configFile); err != nil {
		return err
	}
	return nil
}

// LoadFromReader is a convenience function that creates a ConfigFile object from
// a reader
func LoadFromReader(configData io.Reader) (*ConfigFile, error) {
	configFile := ConfigFile{}
	err := configFile.LoadFromReader(configData)
	return &configFile, err
}

// Load reads the configuration files in the given directory, and sets up
// the auth config information and return values.
// FIXME: use the internal golang config parser
func Load(filename string) (*ConfigFile, error) {
	configFile := ConfigFile{}

	if _, err := os.Stat(filename); err == nil {
		file, err := os.Open(filename)
		if err != nil {
			return &configFile, err
		}
		defer file.Close()
		err = configFile.LoadFromReader(file)
		return &configFile, err
	} else {
		return &configFile, err
	}
	return &configFile, nil
}
