package app

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type AppConfiguration struct {
	StaticDir       string
	TemplatesDir    string
	Database        string
	Debug           bool
	CookieAuthKey   string
	CookieEncKey    string
	SessionStoreDir string
}

var Configuration AppConfiguration

func LoadConfiguration(filename string) *AppConfiguration {
	log.Print("Loading configuration file ", filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Error: %s\n", err.Error())
		return nil
	}
	err = json.Unmarshal(file, &Configuration)
	if err != nil {
		log.Fatal("Error: %s\n", err.Error())
		return nil
	}
	err = json.Unmarshal(file, Configuration)
	return &Configuration
}
