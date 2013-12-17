package users

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"net/http"
)

var subRouter *mux.Router

func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyLogged(mainPageHandler)).Name("net-index")
	subRouter.HandleFunc("/{page}", app.VerifyLogged(mainPageHandler))
}

type NetPageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Data        string
}

func newNetPageCtx(w http.ResponseWriter, r *http.Request) *NetPageCtx {
	ctx := &NetPageCtx{BasePageContext: app.NewBasePageContext("Network", w, r)}
	ctx.LocalMenu = []app.MenuItem{app.NewMenuItem("IFConfig", "ifconfig"),
		app.NewMenuItem("IPTables", "iptables"),
		app.NewMenuItem("Netstat", "netstat-listen"),
		app.NewMenuItem("Conenctions", "connenctions")}
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
	case "netstat-listen":
		data.Data = h.ReadFromCommand("sudo", "netstat", "-lpn", "--inet")
	case "connenctions":
		data.Data = h.ReadFromCommand("sudo", "netstat", "-pn", "--inet")
	}
	data.CurrentLocalMenuPos = page
	data.CurrentPage = page
	app.RenderTemplate(w, data, "base", "base.tmpl", "log.tmpl", "flash.tmpl")
}
