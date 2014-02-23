package utils

import (
	"encoding/json"
	"io/ioutil"
	l "k.prv/rpimon/helpers/logging"
)

type utility struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

type configuration struct {
	Utils map[string]([]utility) `json:"utils"`
}

var config configuration

// Init utils pages
func loadConfiguration(filename string) error {
	l.Print("pages.utils Loading configuration file: %s ", filename)
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
