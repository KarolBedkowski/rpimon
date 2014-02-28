package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	l "k.prv/rpimon/helpers/logging"
)

type utility struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

type configuration struct {
	Utils map[string]([]*utility) `json:"utils"`
}

var config configuration

// Init utils pages
func loadConfiguration(filename string) error {
	l.Print("modules.utils Loading configuration file: %s ", filename)
	if filename == "" {
		return errors.New("missing configuration")
	}
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		l.Error("pages.utils: error: %s ", err.Error())
		return err
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		l.Error("pages.utils: error: %s", err.Error())
	}
	l.Print("pages.utils Loaded groups: %d ", len(config.Utils))
	return err
}

func saveConfiguration(filename string) error {
	l.Printf("modules.utils.saveConfiguration: Writing configuration to %s\n", filename)
	data, err := json.Marshal(config)
	if err != nil {
		l.Printf("modiles.utils.saveConfiguration: error marshal configuration: %s\n", err)
		return err
	}
	return ioutil.WriteFile(filename, data, 0)
}
