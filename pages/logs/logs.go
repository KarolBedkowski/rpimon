package logs

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"os/exec"
)

var subRouter *mux.Router

func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyLogged(mainPageHandler)).Name("logs-index")
	subRouter.HandleFunc("/{page}", app.VerifyLogged(mainPageHandler))
}

type PageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Data        string
}

func newNetPageCtx(w http.ResponseWriter, r *http.Request) *PageCtx {
	ctx := &PageCtx{BasePageContext: app.NewBasePageContext("logs", w, r)}
	ctx.LocalMenu = []app.MenuItem{app.NewMenuItem("Short", "short"),
		app.NewMenuItem("DMESG", "dmesg"),
		app.NewMenuItem("Syslog", "syslog")}
	ctx.CurrentMainMenuPos = "/logs/"
	return ctx
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	data := newNetPageCtx(w, r)
	vars := mux.Vars(r)
	page, ok := vars["page"]
	if !ok {
		page = "short"
	}
	switch page {
	case "short":
		lines, err := h.ReadFromFileLastLines("/var/log/syslog", 20)
		if err != nil {
			data.Data = err.Error()
		} else {
			data.Data = lines
		}
	case "dmesg":
		data.Data = readFromCommand("dmesg")
	case "syslog":
		lines, err := h.ReadFromFileLastLines("/var/log/syslog", 500)
		if err != nil {
			data.Data = err.Error()
		} else {
			data.Data = lines
		}
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
