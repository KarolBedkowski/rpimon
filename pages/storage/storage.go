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
	subRouter.HandleFunc("/", app.VerifyLogged(mainPageHandler)).Name("storage-index")
	subRouter.HandleFunc("/{page}", app.VerifyLogged(mainPageHandler))
}

type NetPageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Data        string
}

func newNetPageCtx(w http.ResponseWriter, r *http.Request) *NetPageCtx {
	ctx := &NetPageCtx{BasePageContext: app.NewBasePageContext("Storage", w, r)}
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
		data.Data = readFromCommand("df", "-h")
	case "mount":
		data.Data = readFromCommand("sudo", "mount")
	case "devices":
		data.Data = readFromCommand("lsblk")
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
