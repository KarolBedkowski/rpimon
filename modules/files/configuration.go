package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	l "k.prv/rpimon/helpers/logging"
	"path/filepath"
)

type configuration struct {
	BaseDir string
}

var config configuration

// Init utils pages
func loadConfiguration(filename string) error {
	l.Print("pages.files Loading configuration file %s", filename)
	if filename == "" {
		return errors.New("missing configuration")
	}
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		l.Error("pages.files: %s", err.Error())
		return err
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		l.Error("pages.files: %s", err.Error())
		return err
	}

	config.BaseDir, err = filepath.Abs(config.BaseDir)
	if err != nil {
		l.Error("pages.files: error setting absolute base dir ", err.Error())
	}

	return err
}