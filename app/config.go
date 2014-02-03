package app

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// AppConfiguration Main app configuration.
type AppConfiguration struct {
	StaticDir             string
	TemplatesDir          string
	Users                 string
	Debug                 bool
	CookieAuthKey         string
	CookieEncKey          string
	SessionStoreDir       string
	LogFilename           string
	UtilsFilename         string
	MpdHost               string
	BrowserConf           string
	HTTPAddress           string `json:"http_address"`
	HTTPSAddress          string `json:"https_address"`
	SslCert               string
	SslKey                string
	MonitorUpdateInterval int
	Logs                  string
}

// Configuration - main app configuration instance
var Configuration AppConfiguration

// LoadConfiguration from given file
func LoadConfiguration(filename string) *AppConfiguration {
	log.Print("Loading configuration file ", filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("app.LoadConfiguration error: ", err.Error())
		return nil
	}
	err = json.Unmarshal(file, &Configuration)
	if err != nil {
		log.Fatal("app.LoadConfiguration error: ", err.Error())
		return nil
	}
	return &Configuration
}
