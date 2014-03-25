package cfg

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"runtime"
)

type (
	// AppConfiguration Main app configuration.
	AppConfiguration struct {
		StaticDir       string
		TemplatesDir    string
		Users           string
		Debug           bool
		CookieAuthKey   string
		CookieEncKey    string
		SessionStoreDir string
		LogFilename     string
		HTTPAddress     string `json:"http_address"`
		HTTPSAddress    string `json:"https_address"`
		SslCert         string
		SslKey          string
		Monitor         *MonitorConfiguration        `json:"monitor"`
		Modules         map[string]map[string]string `json:"modules"`
	}

	// MonitoredService configure one service to monitor in Monitor module
	MonitoredService struct {
		Port uint32 `json:"port"`
		Name string `json:"name"`
	}

	// MonitoredHost defuine one host to monitor by Monitor module
	MonitoredHost struct {
		Name string `json:"name"`

		Address string `json:"address"`
		// checking method: ping, tcp, http
		Method string `json:"method"`
		// inteval in sec.
		Interval int `json:"interval"`
		// Alarm level: 0=none;
		// when unavailable 1=info, 2=warn, 3=error
		// when available: 11=info, 12=warn, 13=error
		Alarm int `json:"alarm"`
	}

	// MonitorConfiguration hold configuration for Monitor module
	MonitorConfiguration struct {
		UpdateInterval        int                `json:"update_interval"`
		LoadWarning           float64            `json:"load_warning"`
		LoadError             float64            `json:"load_error"`
		RAMUsageWarning       int                `json:"ram_usage_warning"`
		SwapUsageWarning      int                `json:"swap_usage_warning"`
		DefaultFSUsageWarning int                `json:"fs_usage_warning"`
		DefaultFSUsageError   int                `json:"fs_usage_error"`
		CPUTempWarning        int                `json:"cpu_temp_warning"`
		CPUTempError          int                `json:"cpu_temp_error"`
		MonitoredServices     []MonitoredService `json:"monitored_services"`
		CPUFreqFile           string             `json:"cpu_freq_file"`
		CPUTempFile           string             `json:"cpu_temp_file"`
		MonitoredHosts        []MonitoredHost    `json:"monitored_hosts"`
	}
)

// Configuration - main app configuration instance
var Configuration AppConfiguration
var configFilename string

// LoadConfiguration from given file
func LoadConfiguration(filename string) *AppConfiguration {
	log.Print("Loading configuration file ", filename)
	configFilename = filename

	if !loadConfiguration(filename) {
		return nil
	}

	return &Configuration
}

func loadConfiguration(filename string) bool {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Print("Errors: app.LoadConfiguration error: ", err.Error())
		Configuration.loadDefaults()
	} else {
		if err = json.Unmarshal(file, &Configuration); err != nil {
			log.Print("Error: app.LoadConfiguration error: ", err.Error())
			log.Print("Error: Loading default configuration")
			Configuration.loadDefaults()
		}
	}
	Configuration.validate()
	return true
}

// SaveConfiguration write current configuration to json file
func SaveConfiguration() error {
	log.Printf("SaveConfiguration: Writing configuration to %s\n", configFilename)
	data, err := json.Marshal(Configuration)
	if err != nil {
		log.Printf("SaveConfiguration: error marshal configuration: %s\n", err)
		return err
	}
	return ioutil.WriteFile(configFilename, data, 0600)
}

func (ac *AppConfiguration) loadDefaults() {
	ac.StaticDir = "./static"
	ac.TemplatesDir = "./templates"
	ac.Users = "./users.json"
	ac.Debug = true
	ac.CookieAuthKey = "12345678901234567890123456789012"
	ac.CookieEncKey = "12345678901234567890123456789012"
	ac.SessionStoreDir = "./temp"
	ac.LogFilename = "app.log"
	ac.HTTPAddress = ":8000"
	ac.HTTPSAddress = ""
	ac.SslCert = "key.pem"
	ac.SslKey = "cert.pem"
	ac.Monitor = &MonitorConfiguration{}
	ac.Monitor.loadDefaults()
}

func (ac *AppConfiguration) validate() {
	ac.Monitor.validate()
}

func (mc *MonitorConfiguration) loadDefaults() {
	mc.UpdateInterval = 5
	mc.LoadWarning = float64(runtime.NumCPU() * 2)
	mc.LoadError = float64(runtime.NumCPU() * 4)
	mc.RAMUsageWarning = 90
	mc.SwapUsageWarning = 75
	mc.DefaultFSUsageWarning = 90
	mc.DefaultFSUsageError = 95
	mc.CPUTempWarning = 60
	mc.CPUTempError = 80
	mc.CPUFreqFile = "/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_cur_freq"
	mc.CPUTempFile = "/sys/class/thermal/thermal_zone0/temp"
}

func (mc *MonitorConfiguration) validate() {
	if mc.UpdateInterval < 0 {
		mc.UpdateInterval = 5
	}
	if mc.LoadWarning < 0 {
		mc.LoadWarning = float64(runtime.NumCPU() * 2)
	}
	if mc.LoadError < 0 {
		mc.LoadError = mc.LoadWarning * 2
	}
	if mc.RAMUsageWarning < 0 || mc.RAMUsageWarning > 100 {
		mc.RAMUsageWarning = 90
	}
	if mc.SwapUsageWarning < 0 || mc.SwapUsageWarning > 100 {
		mc.SwapUsageWarning = 75
	}
	if mc.DefaultFSUsageWarning < 0 || mc.DefaultFSUsageWarning > 100 {
		mc.DefaultFSUsageWarning = 90
	}
	if mc.DefaultFSUsageError < 0 || mc.DefaultFSUsageError > 100 {
		mc.DefaultFSUsageError = 95
	}
	if mc.DefaultFSUsageError < mc.DefaultFSUsageWarning && mc.DefaultFSUsageError > 0 {
		mc.DefaultFSUsageError = 0
	}
	if mc.CPUTempWarning < 0 {
		mc.CPUTempWarning = 60
	}
	if mc.CPUTempError < 0 {
		mc.CPUTempError = 80
	}
	if mc.CPUTempError < mc.CPUTempWarning && mc.CPUTempError > 0 {
		mc.CPUTempError = 0
	}
}
