package app

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type AppConfiguration struct {
	Root string
}

func (aconf *AppConfiguration) LoadConfiguration(filename string) {
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
