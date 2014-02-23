package app

import (
	"github.com/gorilla/mux"
	l "k.prv/rpimon/helpers/logging"
)

type Privilage struct {
	Name        string
	Description string
}

type Module struct {
	// module internal name
	Name string
	// Module Title
	Title string
	// Module description
	Description string
	// All privilages used by module
	AllPrivilages []Privilage

	// Is module enabled
	Enabled bool
	// filename of module configuration file
	ConfFile string

	// Initialize module (set routes etc)
	Init func(parentRoute *mux.Route, configFilename string, globalConfig *AppConfiguration) bool

	// GetMenu return parent menu idand menu item (with optional submenu)
	GetMenu func(ctx *BasePageContext) (parentId string, menu *MenuItem)

	// GetWarnings return map warning kind -> messages
	GetWarnings func() map[string][]string

	// Shutdown module
	Shutdown func()
}

var registeredModules = make(map[string]*Module)

func Register(module *Module) bool {
	if module.Name == "" {
		module.Name = module.Title
	}
	if module.Init == nil {
		l.Error("Module %v missing Init func.", module)
		return false
	}
	l.Info("Registering module: [%s] %s", module.Name, module.Title)
	registeredModules[module.Title] = module
	return true
}

func InitModules(conf *AppConfiguration, router *mux.Router) {
	for _, module := range registeredModules {
		if mconfig, ok := conf.Modules[module.Name]; !ok || mconfig == nil {
			l.Warn("Missing configuration for %v module", module)
		} else {
			module.Enabled = mconfig.Enabled
			if module.Enabled {
				l.Info("Enabling module %s", module.Name)
				if module.Init(router.PathPrefix("/m/"+module.Name),
					mconfig.ConfigFilename, conf) {
				} else {
					l.Warn("Module %s init error; %#v", module.Name, mconfig)
				}
			}
		}
	}
}

func ShutdownModules() {
	for _, module := range registeredModules {
		if module.Enabled && module.Shutdown != nil {
			module.Shutdown()
		}
	}
}
