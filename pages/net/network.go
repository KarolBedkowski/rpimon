package users

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"net/http"
)

var subRouter *mux.Router

// CreateRoutes for /net
func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(mainPageHandler, "admin")).Name("net-index")
	subRouter.HandleFunc("/{page}", app.VerifyPermission(mainPageHandler, "admin")).Name("net-page")
}

type pageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Data        string
}

var localMenu []app.MenuItem

func createLocalMenu() []app.MenuItem {
	if localMenu == nil {

		localMenu = []app.MenuItem{app.NewMenuItemFromRoute("IFConfig", "net-page", "page", "ifconfig").SetID("ifconfig"),
			app.NewMenuItemFromRoute("IPTables", "net-page", "page", "iptables").SetID("iptables"),
			app.NewMenuItemFromRoute("Netstat", "net-page", "page", "netstat").SetID("netstat"),
			app.NewMenuItemFromRoute("Conenctions", "net-page", "page", "connenctions").SetID("connenctions")}
	}
	return localMenu
}

func newNetPageCtx(w http.ResponseWriter, r *http.Request) *pageCtx {
	ctx := &pageCtx{BasePageContext: app.NewBasePageContext("Network", w, r)}
	ctx.LocalMenu = createLocalMenu()
	ctx.CurrentMainMenuPos = "/net/"
	return ctx
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	data := newNetPageCtx(w, r)
	vars := mux.Vars(r)
	page, ok := vars["page"]
	if !ok {
		page = "ifconfig"
	}
	switch page {
	case "ifconfig":
		data.Data = h.ReadFromCommand("ip", "addr")
	case "iptables":
		data.Data = h.ReadFromCommand("sudo", "iptables", "-L", "-vn")
	case "netstat":
		data.Data = h.ReadFromCommand("sudo", "netstat", "-lpn", "--inet")
	case "connenctions":
		data.Data = h.ReadFromCommand("sudo", "netstat", "-pn", "--inet")
	}
	data.CurrentLocalMenuPos = page
	data.CurrentPage = page
	app.RenderTemplate(w, data, "base", "base.tmpl", "log.tmpl", "flash.tmpl")
}
