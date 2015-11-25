package context

import (
	"fmt"
	"github.com/gorilla/mux"
	"k.prv/rpimon/cfg"
	l "k.prv/rpimon/logging"
	"sort"
)

// Privilege used for modules
type Privilege struct {
	Name        string
	Description string
}

// WarningsStruct holds current warnings, errors and informations.
type WarningsStruct struct {
	Warnings []string
	Errors   []string
	Infos    []string
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
	GetWarnings func() *WarningsStruct

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
	ConfigurePageURL string

	// All privileges used by module
	AllPrivilages []Privilege
}

var (
	registeredModules = make(map[string]*Module)
	appRouter         *mux.Router
	// AllPrivilages privilages defined in all modules
	AllPrivilages = make(map[string]Privilege)
)

// RegisterModule register given module for later usage
func RegisterModule(module *Module) bool {
	if module.Init == nil {
		l.Error(fmt.Sprintf("Module %v missing Init func.", module))
		return false
	}
	l.Info("Registering module: [%s] %s", module.Name, module.Title)
	registeredModules[module.Name] = module
	return true
}

// GetModules return all registered modules
func GetModules() map[string]*Module {
	return registeredModules
}

// InitModules initialize and enable all modules
func InitModules(conf *cfg.AppConfiguration, router *mux.Router) {
	appRouter = router
	for _, module := range registeredModules {
		module.enable(module.Internal || module.GetConfiguration()["enabled"] == "yes")
		if module.AllPrivilages != nil {
			for _, priv := range module.AllPrivilages {
				if _, found := AllPrivilages[priv.Name]; !found {
					AllPrivilages[priv.Name] = priv
					l.Debug("InitModules: add privilage %v", priv)
				}
			}
		}
	}
}

// ShutdownModules shutdown all enabled modules; call Shutdown method for modules.
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

// GetModulesList returns all registered modules as sorted by title list.
func GetModulesList() (modules []*Module) {
	modules = make([]*Module, 0, len(registeredModules))
	for _, module := range registeredModules {
		modules = append(modules, module)
	}
	sort.Sort(ModulesByTitle(modules))
	return
}

// GetModule return module by name
func GetModule(name string) (module *Module) {
	return registeredModules[name]
}

// Enabled check is module enabled
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
	m.SaveConfiguration(mconfig)
}

// GetConfiguration get configuration for current module; load default if not exists
func (m *Module) GetConfiguration() (conf map[string]string) {
	conf = map[string]string{
		"enabled": "",
	}
	for key, val := range m.Defaults {
		conf[key] = val
	}

	if mconfig, ok := cfg.Configuration.Modules[m.Name]; ok && mconfig != nil {
		for k, v := range mconfig {
			conf[k] = v
		}
	} else {
		if m.NeedConfiguration {
			l.Warn("Missing configuration for %v module; loading defaults - module is disabled", m.Name)
		} else {
			conf["enabled"] = "yes"
		}
	}
	return conf
}

// SaveConfiguration update app configuration file for given module
func (m *Module) SaveConfiguration(conf map[string]string) {
	if cfg.Configuration.Modules == nil {
		cfg.Configuration.Modules = map[string]map[string]string{}
	}
	cfg.Configuration.Modules[m.Name] = conf
}

// EnableModule enable or disable module by name.
func EnableModule(name string, enabled bool) {
	if module, ok := registeredModules[name]; ok {
		module.enable(enabled)
	} else {
		l.Warn("EnableModule wrong module %s", name)
		l.Debug("%v", registeredModules)
	}
}

// SORTING Modules ITEMS

// ModulesByTitle type for sorting Modules by title
type ModulesByTitle []*Module

func (s ModulesByTitle) Len() int           { return len(s) }
func (s ModulesByTitle) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ModulesByTitle) Less(i, j int) bool { return s[i].Title < s[j].Title }

// GetWarnings from enabled modules
func GetWarnings() *WarningsStruct {
	w := &WarningsStruct{}
	for _, mod := range registeredModules {
		if mod.Enabled() && mod.GetWarnings != nil {
			if mw := mod.GetWarnings(); mw != nil {
				w.Infos = append(w.Infos, mw.Infos...)
				w.Warnings = append(w.Warnings, mw.Warnings...)
				w.Errors = append(w.Errors, mw.Errors...)
			}
		}
	}
	return w
}
