package modules

import (
	"k.prv/rpimon/app"
	l "k.prv/rpimon/helpers/logging"
)

type Privilage struct {
	Name        string
	Description string
}

type Module struct {
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

	// GetMenu return parent menu idand menu item (with optional submenu)
	GetMenu func(ctx *app.BasePageContext) (parentId string, menu *app.MenuItem)

	// GetWarnings return map warning kind -> messages
	GetWarnings func() map[string][]string
}

var Modules = make(map[string]*Module)

func Register(module *Module) bool {
	l.Info("Registering module: %s", module.Title)
	Modules[module.Title] = module
	app.RegisterMenuItem(module.GetMenu)

	return true
}
