package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type utility struct {
	Name    string `json:name`
	Command string `json:command`
}

type configuration struct {
	Utils map[string]([]utility) `json:utils"`
}

var config configuration

// Init utils pages
func Init(filename string) error {
	log.Print("pages.utils Loading configuration file ", filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Error: ", err.Error())
		return err
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatal("Error: ", err.Error())
	}
	log.Print("pages.utils Loaded groups ", len(config.Utils))
	return err
}
