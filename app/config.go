package app

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type AppConfiguration struct {
	StaticDir       string
	TemplatesDir    string
	Users           string
	Debug           bool
	CookieAuthKey   string
	CookieEncKey    string
	SessionStoreDir string
	LogFilename     string
	UtilsFilename   string
}

var Configuration AppConfiguration

func LoadConfiguration(filename string) *AppConfiguration {
	log.Print("Loading configuration file ", filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Error: ", err.Error())
		return nil
	}
	err = json.Unmarshal(file, &Configuration)
	if err != nil {
		log.Fatal("Error: ", err.Error())
		return nil
	}
	return &Configuration
}
