package logs

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"net/http"
)

// CreateRoutes for /users
func CreateRoutes(parentRoute *mux.Route) {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(mainPageHandler, "admin")).Name("other-index")
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	page := r.FormValue("sec")
	if page == "" {
		page = "acpi"
	}
	data := app.NewSimpleDataPageCtx(w, r, "Other", "", nil)
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
