package app

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
)

type (
	// AppConfiguration Main app configuration.
	AppConfiguration struct {
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
		Notepad               string               `json:"notepad"`
		Monitor               MonitorConfiguration `json:"monitor"`
	}
	// monitor configuration
	MonitorConfiguration struct {
		LoadWarning           float64           `json:"load_warning"`
		LoadError             float64           `json:"load_error"`
		RamUsageWarning       int               `json:"ram_usage_warning"`
		SwapUsageWarning      int               `json:"swap_usage_warning"`
		DefaultFSUsageWarning int               `json:"fs_usage_warning"`
		DefaultFSUsageError   int               `json:"fs_usage_error"`
		CPUTempWarning        int               `json:"cpu_temp_warning"`
		CPUTempError          int               `json:"cpu_temp_error"`
		MonitoredServices     map[string]string `json:"monitored_services"`
	}
)

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

	if err = json.Unmarshal(file, &Configuration); err != nil {
		log.Fatal("app.LoadConfiguration error: ", err.Error())
		return nil
	}

	Configuration.Notepad, _ = filepath.Abs(Configuration.Notepad)

	if Configuration.Monitor.LoadWarning == 0 {
		Configuration.Monitor.LoadWarning = float64(runtime.NumCPU())
	}
	if Configuration.Monitor.LoadError == 0 {
		Configuration.Monitor.LoadError = float64(runtime.NumCPU() * 2)
	}
	if Configuration.Monitor.RamUsageWarning == 0 {
		Configuration.Monitor.RamUsageWarning = 90
	}
	if Configuration.Monitor.SwapUsageWarning == 0 {
		Configuration.Monitor.SwapUsageWarning = 75
	}
	if Configuration.Monitor.DefaultFSUsageWarning == 0 {
		Configuration.Monitor.DefaultFSUsageWarning = 90
	}
	if Configuration.Monitor.DefaultFSUsageError == 0 {
		Configuration.Monitor.DefaultFSUsageError = 95
	}
	if Configuration.Monitor.CPUTempWarning == 0 {
		Configuration.Monitor.CPUTempWarning = 60
	}
	if Configuration.Monitor.CPUTempError == 0 {
		Configuration.Monitor.CPUTempError = 80
	}

	return &Configuration
}
