package network

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"net/http"
)

// NFSModule information
var NFSModule = &app.Module{
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
	subRouter.HandleFunc("/", app.TimeoutHandler(app.SecContext(nfsPageHandler, "Network - NFS", "admin"), 5)).Name("m-net-nfs")
	return true
}

func nfsGetMenu(ctx *app.BaseCtx) (parentID string, menu *app.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		return "", nil
	}

	menu = app.NewMenuItemFromRoute("NFS", "m-net-nfs")
	return "m-net", menu
}

func nfsPageHandler(r *http.Request, ctx *app.BaseCtx) {
	page := r.FormValue("sec")
	if page == "" {
		page = "stat"
	}
	data := &app.DataPageCtx{BaseCtx: ctx}
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
	data.Tabs = []*app.MenuItem{
		app.NewMenuItemFromRoute("NFSstat", "m-net-nfs").AddQuery("?sec=stat").SetActve(page == "stat"),
		app.NewMenuItemFromRoute("exportfs", "m-net-nfs").AddQuery("?sec=exportfs").SetActve(page == "exportfs"),
	}
	ctx.RenderStd(data, "data.tmpl", "tabs.tmpl")
}
