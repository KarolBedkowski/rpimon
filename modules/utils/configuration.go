package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	l "k.prv/rpimon/logging"
	"sort"
)

type utility struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

type configuration struct {
	Utils map[string]([]*utility) `json:"utils"`
}

var config configuration

type utilitiesByName []*utility

func (s utilitiesByName) Len() int           { return len(s) }
func (s utilitiesByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s utilitiesByName) Less(i, j int) bool { return s[i].Name < s[j].Name }

// Init utils pages
func loadConfiguration(filename string) error {
	l.Info("modules.utils Loading configuration file: %s ", filename)
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
	l.Info("pages.utils Loaded groups: %d ", len(config.Utils))
	return err
}

func saveConfiguration(filename string) error {
	l.Info("modules.utils.saveConfiguration: Writing configuration to %s\n", filename)
	for _, utils := range config.Utils {
		sort.Sort(utilitiesByName(utils))
	}
	data, err := json.Marshal(config)
	if err != nil {
		l.Info("modiles.utils.saveConfiguration: error marshal configuration: %s\n", err)
		return err
	}
	return ioutil.WriteFile(filename, data, 0600)
}
