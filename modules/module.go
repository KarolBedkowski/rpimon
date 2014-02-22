package modules

import (
	"k.prv/rpimon/app"
	l "k.prv/rpimon/helpers/logging"
)

type Privilage struct {
	Name        string
	Description string
}

type ModuleInfo struct {
	Title         string
	Description   string
	AllPrivilages []Privilage
}

type Module interface {
	// GetMenu return parent menu idand menu item (with optional submenu)
	GetMenu(ctx *app.BasePageContext) (parentId string, menu *app.MenuItem)

	// GetWarnings return map warning kind -> messages
	GetWarnings() map[string][]string

	// GetInfo about module
	GetInfo() *ModuleInfo
}

var Modules = make(map[string]Module)

func Register(module Module) bool {
	info := module.GetInfo()
	l.Info("Registering module: %s", info.Title)
	Modules[info.Title] = module
	app.RegisterMenuItem(module.GetMenu)

	return true
}
