package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
)

type configuration struct {
	BaseDir string
}

var config configuration

// Init utils pages
func Init(filename string) error {
	log.Print("pages.files Loading configuration file ", filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Error: ", err.Error())
		return err
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatal("Error: ", err.Error())
	}

	config.BaseDir, err = filepath.Abs(config.BaseDir)
	if err != nil {
		log.Fatal("error setting absolute base dir ", err.Error())
	}

	return err
}
