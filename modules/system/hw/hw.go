package systemhw

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	h "k.prv/rpimon/helpers"
	"net/http"
)

// Module information
var Module = &context.Module{
	Name:          "system-hw",
	Title:         "System - Hardware",
	Description:   "",
	AllPrivilages: nil,
	Init:          initModule,
	GetMenu:       getMenu,
}

// CreateRoutes for /users
func initModule(parentRoute *mux.Route) bool {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", context.SecContext(mainPageHandler, "Hardware", "admin")).Name("sys-hw-index")
	return true
}

func getMenu(ctx *context.BaseCtx) (parentID string, menu *context.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		return "", nil
	}
	menu = app.NewMenuItemFromRoute("Hardware", "sys-hw-index").SetID("sys-hw-index").SetIcon("glyphicon glyphicon-cog").SetSortOrder(90)
	return "system", menu
}

func mainPageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BaseCtx) {
	page := r.FormValue("sec")
	if page == "" {
		page = "acpi"
	}
	data := &context.DataPageCtx{BaseCtx: bctx}
	data.Header1 = "Hardware"
	data.Tabs = []*context.MenuItem{
		app.NewMenuItemFromRoute("ACPI", "other-index").AddQuery("?sec=acpi").SetActve(page == "acpi"),
		app.NewMenuItemFromRoute("Sensors", "other-index").AddQuery("?sec=sensors").SetActve(page == "sensors"),
		app.NewMenuItemFromRoute("lscpu", "other-index").AddQuery("?sec=lscpu").SetActve(page == "lscpu"),
		app.NewMenuItemFromRoute("lsusb", "other-index").AddQuery("?sec=lsusb").SetActve(page == "lsusb"),
		app.NewMenuItemFromRoute("lspci", "other-index").AddQuery("?sec=lspci").SetActve(page == "lspci"),
	}
	switch page {
	case "acpi":
		data.Data = h.ReadCommand("acpi", "-V", "-i")
		data.Header2 = "ACPI"
	case "sensors":
		data.Data = h.ReadCommand("sensors")
		data.Header2 = "Sensors"
	case "lscpu":
		data.Data = h.ReadCommand("lscpu")
		data.Header2 = "lscpu"
	case "lsusb":
		data.Data = h.ReadCommand("lsusb")
		data.Header2 = "lsusb"
	case "lspci":
		data.Data = h.ReadCommand("lspci")
		data.Header2 = "lspci"
	}
	data.SetMenuActive("sys-hw-index")
	app.RenderTemplateStd(w, data, "data.tmpl", "tabs.tmpl")
}
