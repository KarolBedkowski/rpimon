package preferences

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	"k.prv/rpimon/modules/preferences/modules"
	"k.prv/rpimon/modules/preferences/users"
)

// Module information
var Module *context.Module

func init() {
	Module = &context.Module{
		Name:          "preferences",
		Title:         "Preferences",
		Description:   "System preferences",
		AllPrivilages: nil,
		Init:          initModule,
		GetMenu:       getMenu,
		Internal:      true,
		Configurable:  false,
	}
}

// CreateRoutes for /pages
func initModule(parentRoute *mux.Route) bool {
	subRouter := parentRoute.Subrouter()
	modules.CreateRoutes(subRouter.PathPrefix("/pref/modules"))
	users.CreateRoutes(subRouter.PathPrefix("/pref/users"))
	return true
}
func getMenu(ctx *context.BasePageContext) (parentID string, menu *context.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		return "", nil
	}

	menu = context.NewMenuItem("Preferences", "preferences").SetSortOrder(999).SetIcon("glyphicon glyphicon-wrench")
	// Preferences
	if context.CheckPermission(ctx.CurrentUserPerms, "admin") {
		//TODO: named routes
		menu.AddChild(context.NewMenuItem("Modules", app.GetNamedURL("m-pref-modules-index")).SetID("p-modules"))
		menu.AddChild(context.NewMenuItem("Users", app.GetNamedURL("m-pref-users-index")).SetID("p-users"))
	}
	if ctx.CurrentUser != "" {
		menu.AddChild(context.NewMenuItem("Profile", app.GetNamedURL("m-pref-user-profile")).SetID("p-user-profile"))
	}

	return "", menu
}
