package preferences

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	"k.prv/rpimon/modules/preferences/modules"
	"k.prv/rpimon/modules/preferences/users"
)

// Module information
var Module *app.Module

func init() {
	Module = &app.Module{
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
func getMenu(ctx *app.BaseCtx) (parentID string, menu *app.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		return "", nil
	}

	menu = app.NewMenuItem("Preferences", "preferences").SetSortOrder(999).SetIcon("glyphicon glyphicon-cog")
	// Preferences
	if app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		//TODO: named routes
		menu.AddChild(app.NewMenuItem("Modules", app.GetNamedURL("m-pref-modules-index")).SetID("m-modules"))
		menu.AddChild(app.NewMenuItem("Users", app.GetNamedURL("m-pref-users-index")).SetID("m-users"))
	}
	if ctx.CurrentUser != "" {
		menu.AddChild(app.NewMenuItem("Profile", app.GetNamedURL("m-pref-user-profile")).SetID("m-user-profile"))
	}

	return "", menu
}
