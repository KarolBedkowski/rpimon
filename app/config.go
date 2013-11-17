package app

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type AppConfiguration struct {
	Root         string
	StaticDir    string
	TemplatesDir string
	Database     string
}

func (aconf *AppConfiguration) LoadConfiguration(filename string) {
	log.Print("Loading configuration file ", filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Error: %s\n", err.Error())
		return
	}
	err = json.Unmarshal(file, aconf)
	if err != nil {
		log.Fatal("Error: %s\n", err.Error())
		return
	}
}
