package users

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"net/http"
)

var subRouter *mux.Router

// CreateRoutes for /storage
func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(mainPageHandler, "admin")).Name("storage-index")
	subRouter.HandleFunc("/{page}", app.VerifyPermission(mainPageHandler, "admin")).Name("storage-page")
}

type pageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Data        string
}

var localMenu []app.MenuItem

func createLocalMenu() []app.MenuItem {
	if localMenu == nil {
		localMenu = []app.MenuItem{app.NewMenuItemFromRoute("Disk Free", "storage-page", "page", "diskfree").SetID("diskfree"),
			app.NewMenuItemFromRoute("Mount", "storage-page", "page", "mount").SetID("mount"),
			app.NewMenuItemFromRoute("Devices", "storage-page", "page", "devices").SetID("devices")}
	}
	return localMenu
}

func newNetPageCtx(w http.ResponseWriter, r *http.Request) *pageCtx {
	ctx := &pageCtx{BasePageContext: app.NewBasePageContext("Storage", "storage", w, r)}
	ctx.LocalMenu = createLocalMenu()
	return ctx
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	data := newNetPageCtx(w, r)
	vars := mux.Vars(r)
	page, ok := vars["page"]
	if !ok {
		page = "diskfree"
	}
	switch page {
	case "diskfree":
		data.Data = h.ReadFromCommand("df", "-h")
	case "mount":
		data.Data = h.ReadFromCommand("sudo", "mount")
	case "devices":
		data.Data = h.ReadFromCommand("lsblk")
	}
	data.CurrentLocalMenuPos = page
	data.CurrentPage = page
	app.RenderTemplate(w, data, "base", "base.tmpl", "log.tmpl", "flash.tmpl")
}
