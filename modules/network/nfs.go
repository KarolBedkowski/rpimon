package network

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	h "k.prv/rpimon/helpers"
	"net/http"
)

// NFSModule information
var NFSModule = &context.Module{
	Name:          "network-nfs",
	Title:         "Network - NFS",
	Description:   "Network - NFS",
	AllPrivilages: nil,
	Init:          initNFSModule,
	GetMenu:       nfsGetMenu,
}

// initNFSModule initialize module
func initNFSModule(parentRoute *mux.Route) bool {
	// todo register modules
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", context.SecContext(nfsPageHandler, "Network - NFS", "admin")).Name("m-net-nfs")
	return true
}

func nfsGetMenu(ctx *context.BaseCtx) (parentID string, menu *context.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		return "", nil
	}

	menu = app.NewMenuItemFromRoute("NFS", "m-net-nfs")
	return "m-net", menu
}

func nfsPageHandler(w http.ResponseWriter, r *http.Request, ctx *context.BaseCtx) {
	page := r.FormValue("sec")
	if page == "" {
		page = "stat"
	}
	data := &context.DataPageCtx{BaseCtx: ctx}
	data.SetMenuActive("m-net-nfs")
	data.Header1 = "NFS"
	switch page {
	case "exportfs":
		data.Header2 = "Listen"
		data.Data = h.ReadCommand("sudo", "exportfs", "-v")
	case "stat":
		data.Header2 = "Connections"
		data.Data = h.ReadCommand("nfsstat")
	}
	data.Tabs = []*context.MenuItem{
		app.NewMenuItemFromRoute("NFSstat", "m-net-nfs").AddQuery("?sec=stat").SetActve(page == "stat"),
		app.NewMenuItemFromRoute("exportfs", "m-net-nfs").AddQuery("?sec=exportfs").SetActve(page == "exportfs"),
	}
	app.RenderTemplateStd(w, data, "data.tmpl", "tabs.tmpl")
}
