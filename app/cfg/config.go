package cfg

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"runtime"
	"time"
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
		HTTPAddress           string `json:"http_address"`
		HTTPSAddress          string `json:"https_address"`
		SslCert               string
		SslKey                string
		MonitorUpdateInterval int
		Monitor               MonitorConfiguration         `json:"monitor"`
		Modules               map[string]map[string]string `json:"modules"`
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
		CPUFreqFile           string            `json:"cpu_freq_file"`
		CPUTempFile           string            `json:"cpu_temp_file"`
	}
)

// Configuration - main app configuration instance
var Configuration AppConfiguration
var quitReloaderChan = make(chan string)
var configFilename string

// LoadConfiguration from given file
func LoadConfiguration(filename string) *AppConfiguration {
	log.Print("Loading configuration file ", filename)
	configFilename = filename

	if !loadConfiguration(filename) {
		return nil
	}

	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				//log.Print("Reloading configuration")
				loadConfiguration(filename)
			case <-quitReloaderChan:
				ticker.Stop()
				return
			}
		}
	}()

	return &Configuration
}

func CloseConf() {
	quitReloaderChan <- "END"
}

func loadConfiguration(filename string) bool {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("app.LoadConfiguration error: ", err.Error())
		return false
	}

	if err = json.Unmarshal(file, &Configuration); err != nil {
		log.Fatal("app.LoadConfiguration error: ", err.Error())
		return false
	}

	if Configuration.Monitor.LoadWarning == 0 {
		Configuration.Monitor.LoadWarning = float64(runtime.NumCPU() * 2)
	}
	if Configuration.Monitor.LoadError == 0 {
		Configuration.Monitor.LoadError = float64(runtime.NumCPU() * 4)
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
	return true
}

func SaveConfiguration() error {
	log.Printf("SaveConfiguration: Writing configuration to %s\n", configFilename)
	data, err := json.Marshal(Configuration)
	if err != nil {
		log.Printf("SaveConfiguration: error marshal configuration: %s\n", err)
		return err
	}
	return ioutil.WriteFile(configFilename, data, 0)
}
