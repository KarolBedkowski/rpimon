package network

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"net/http"
)

// SambaModule - module information
var SambaModule = &app.Module{
	Name:          "network-smb",
	Title:         "Network - SAMBA",
	Description:   "Network - SAMBA",
	AllPrivilages: nil,
	Init:          initSambaModule,
	GetMenu:       smbGetMenu,
}

func initSambaModule(parentRoute *mux.Route) bool {
	// todo register modules
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/",
		app.TimeoutHandler(app.SecContext(sambaPageHandler, "Network - Samba", "admin"), 5)).
		Name("m-net-samba")
	return true
}

func smbGetMenu(ctx *app.BaseCtx) (parentID string, menu *app.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		return "", nil
	}

	menu = app.NewMenuItemFromRoute("Samba", "m-net-samba")
	return "m-net", menu
}

func sambaPageHandler(w http.ResponseWriter, r *http.Request, ctx *app.BaseCtx) {
	data := &app.DataPageCtx{BaseCtx: ctx}
	data.SetMenuActive("m-net-samba")
	data.Header1 = "Samba"
	data.Data = h.ReadCommand("sudo", "smbstatus")
	app.RenderTemplateStd(w, data, "data.tmpl")
}
