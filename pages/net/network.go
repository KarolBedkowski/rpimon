package users

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"os/exec"
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
		data.Data = readFromCommand("ip", "addr")
	case "iptables":
		data.Data = readFromCommand("sudo", "iptables", "-L", "-vn")
	case "netstat-listen":
		data.Data = readFromCommand("sudo", "netstat", "-lpn", "--inet")
	case "connenctions":
		data.Data = readFromCommand("sudo", "netstat", "-pn", "--inet")
	}
	data.CurrentLocalMenuPos = page
	data.CurrentPage = page
	app.RenderTemplate(w, data, "base", "base.tmpl", "log.tmpl", "flash.tmpl")
}

func readFromCommand(name string, arg ...string) string {
	out, err := exec.Command(name, arg...).Output()
	if err != nil {
		l.Warn("readFromCommand Error", name, arg, err)
		return err.Error()
	}
	return string(out)
}
