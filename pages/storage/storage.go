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
	subRouter.HandleFunc("/", app.VerifyLogged(mainPageHandler)).Name("storage-index")
	subRouter.HandleFunc("/{page}", app.VerifyLogged(mainPageHandler))
}

type pageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Data        string
}

func newNetPageCtx(w http.ResponseWriter, r *http.Request) *pageCtx {
	ctx := &pageCtx{BasePageContext: app.NewBasePageContext("Storage", w, r)}
	ctx.LocalMenu = []app.MenuItem{app.NewMenuItem("Disk Free", "diskfree"),
		app.NewMenuItem("Mount", "mount"),
		app.NewMenuItem("Devices", "devices")}
	ctx.CurrentMainMenuPos = "/storage/"
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
