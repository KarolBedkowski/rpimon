package modules

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
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
	Init func(parentRoute *mux.Route, configFilename string, globalConfig *app.AppConfiguration) bool

	// GetMenu return parent menu idand menu item (with optional submenu)
	GetMenu func(ctx *app.BasePageContext) (parentId string, menu *app.MenuItem)

	// GetWarnings return map warning kind -> messages
	GetWarnings func() map[string][]string

	// Shutdown module
	Shutdown func()
}

var Modules = make(map[string]*Module)

func Register(module *Module) bool {
	if module.Name == "" {
		module.Name = module.Title
	}
	if module.Init == nil {
		l.Error("Module %v missing Init func.", module)
		return false
	}
	l.Info("Registering module: [%s] %s", module.Name, module.Title)
	Modules[module.Title] = module
	return true
}

func InitModules(conf *app.AppConfiguration, router *mux.Router) {
	for _, module := range Modules {
		if mconfig, ok := conf.Modules[module.Name]; !ok || mconfig == nil {
			l.Warn("Missing configuration for %v module", module)
		} else {
			module.Enabled = mconfig.Enabled
			if module.Enabled {
				l.Info("Enabling module %s", module.Name)
				if module.Init(router.PathPrefix("/m/"+module.Name),
					mconfig.ConfigFilename, conf) {
					if module.GetMenu != nil {
						app.RegisterMenuProvider(module.GetMenu)
					}
				}
			}
		}
	}
}

func ShutdownModules() {
	for _, module := range Modules {
		if module.Enabled && module.Shutdown != nil {
			module.Shutdown()
		}
	}
}
