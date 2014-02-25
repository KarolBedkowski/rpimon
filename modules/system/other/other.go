package logs

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"net/http"
)

func GetModule() *app.Module {
	return &app.Module{
		Name:          "system-other",
		Title:         "Other",
		Description:   "",
		AllPrivilages: nil,
		Init:          initModule,
		GetMenu:       getMenu,
	}
}

// CreateRoutes for /users
func initModule(parentRoute *mux.Route, conf *app.ModuleConf, gconf *app.AppConfiguration) bool {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.HandleWithContextSec(mainPageHandler, "Other", "admin")).Name("other-index")
	return true
}

func getMenu(ctx *app.BasePageContext) (parentId string, menu *app.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		return "", nil
	}
	menu = app.NewMenuItemFromRoute("Other", "other-index").SetID("other").SetIcon("glyphicon glyphicon-cog").SetSortOrder(999)
	return "system", menu
}

func mainPageHandler(w http.ResponseWriter, r *http.Request, bctx *app.BasePageContext) {
	page := r.FormValue("sec")
	if page == "" {
		page = "acpi"
	}
	data := &app.SimpleDataPageCtx{BasePageContext: bctx}
	data.Header1 = "Other"
	data.Tabs = []*app.MenuItem{
		app.NewMenuItemFromRoute("ACPI", "other-index").AddQuery("?sec=acpi").SetActve(page == "acpi"),
		app.NewMenuItemFromRoute("Sensors", "other-index").AddQuery("?sec=sensors").SetActve(page == "sensors"),
	}
	data.SetMenuActive("other")
	switch page {
	case "acpi":
		data.Data = h.ReadCommand("acpi", "-V", "-i")
		data.Header2 = "ACPI"
	case "sensors":
		data.Data = h.ReadCommand("sensors")
		data.Header2 = "Sensors"
	}
	app.RenderTemplateStd(w, data, "data.tmpl", "tabs.tmpl")
}
