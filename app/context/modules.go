package context

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app/cfg"
	l "k.prv/rpimon/helpers/logging"
	"sort"
)

// Privilege used for modules
type Privilege struct {
	Name        string
	Description string
}

// Module definition structore
type Module struct {
	// module internal name
	Name string
	// Module Title
	Title string
	// Module description
	Description string

	// internal / always enabled
	Internal bool

	// Is module initialized
	initialized bool

	// Initialize module (set routes etc)
	Init func(parentRoute *mux.Route) bool

	// GetMenu return parent menu id and menu item (with optional submenu)
	GetMenu MenuGenerator

	// GetWarnings return map warning kind -> messages
	GetWarnings func() map[string][]string

	// Shutdown module
	Shutdown func()

	// Configuration
	// Is module allow configuration
	Configurable bool
	// Is module requred configuration (except defaults) to start
	NeedConfiguration bool
	// filename of module configuration file
	ConfFile string
	//default configuration
	Defaults map[string]string
	// Page href for custom configuration page
	ConfigurePageUrl string

	// All privileges used by module
	AllPrivilages []Privilege
}

var (
	registeredModules = make(map[string]*Module)
	appRouter         *mux.Router
)

// RegisterModule register given module for later usage
func RegisterModule(module *Module) bool {
	if module.Init == nil {
		l.Error("Module %v missing Init func.", module)
		return false
	}
	l.Info("Registering module: [%s] %s", module.Name, module.Title)
	registeredModules[module.Name] = module
	return true
}

func GetModules() map[string]*Module {
	return registeredModules
}

// InitModules initialize and enable all modules
func InitModules(conf *cfg.AppConfiguration, router *mux.Router) {
	appRouter = router
	for _, module := range registeredModules {
		module.enable(module.Internal || module.GetConfiguration()["enabled"] == "yes")
	}
}

// ShutdownModules
func ShutdownModules() {
	for _, module := range registeredModules {
		if module.Enabled() && module.Shutdown != nil {
			module.Shutdown()
		}
	}
}

// IsModuleAvailable return true when given module is loaded & enable.
func IsModuleAvailable(name string) bool {
	if module, ok := registeredModules[name]; ok {
		return module.Enabled()
	}
	return false
}

func GetModulesList() (modules []*Module) {
	modules = make([]*Module, 0, len(registeredModules))
	for _, module := range registeredModules {
		modules = append(modules, module)
	}
	sort.Sort(ModulesByTitle(modules))
	return
}

func GetModule(name string) (module *Module) {
	return registeredModules[name]
}

func (m *Module) Enabled() (enabled bool) {
	return m.Internal || m.GetConfiguration()["enabled"] == "yes"
}

func (m *Module) enable(enabled bool) {
	l.Debug("enable module %s %v", m.Name, enabled)
	mconfig := m.GetConfiguration()
	mconfig["enabled"] = ""
	if enabled {
		mconfig["enabled"] = "yes"
		if !m.initialized {
			m.initialized = m.Init(appRouter.PathPrefix("/m/" + m.Name))
			if !m.initialized {
				l.Warn("Module %s init error; %#v", m.Name)
				return
			}
		}
	}
}

func (m *Module) GetConfiguration() (conf map[string]string) {
	if mconfig, ok := cfg.Configuration.Modules[m.Name]; ok && mconfig != nil {
		// configuration exists; add missing from defaults
		for key, val := range m.Defaults {
			if _, found := mconfig[key]; !found {
				mconfig[val] = val
			}
		}
		return mconfig
	}
	conf = map[string]string{
		"enabled": "",
	}
	for key, val := range m.Defaults {
		conf[key] = val
	}
	if m.NeedConfiguration {
		l.Warn("Missing configuration for %v module; loading defaults - module is disabled", m.Name)
	} else {
		l.Info("Missing configuration for %v module; loading defaults", m.Name)
		conf["enabled"] = "yes"
	}
	return conf
}

func SetModuleEnabled(name string, enabled bool) {
	if module, ok := registeredModules[name]; ok {
		module.enable(enabled)
	} else {
		l.Warn("SetModuleEnabled wrong module %s", name)
		l.Debug("%v", registeredModules)
	}
}

// SORTING Modules ITEMS

type ModulesByTitle []*Module

func (s ModulesByTitle) Len() int           { return len(s) }
func (s ModulesByTitle) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ModulesByTitle) Less(i, j int) bool { return s[i].Title < s[j].Title }
