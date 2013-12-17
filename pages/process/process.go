package process

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"strings"
)

var subRouter *mux.Router

func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyLogged(mainPageHandler)).Name("process-index")
	subRouter.HandleFunc("/services", app.VerifyLogged(servicesPageHangler))
	subRouter.HandleFunc("/services/{service}/{action}", app.VerifyLogged(serviceActionPageHandler))
	subRouter.HandleFunc("/{page}", app.VerifyLogged(mainPageHandler))
}

type PageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Data        string
}

func newNetPageCtx(w http.ResponseWriter, r *http.Request) *PageCtx {
	ctx := &PageCtx{BasePageContext: app.NewBasePageContext("Process", w, r)}
	ctx.LocalMenu = []app.MenuItem{app.NewMenuItem("PS AXL", "psaxl"),
		app.NewMenuItem("TOP", "top"),
		app.NewMenuItem("Services", "services")}
	ctx.CurrentMainMenuPos = "/process/"
	return ctx
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	data := newNetPageCtx(w, r)
	vars := mux.Vars(r)
	page, ok := vars["page"]
	if !ok {
		page = "psaxl"
	}
	switch page {
	case "psaxl":
		data.Data = h.ReadFromCommand("ps", "axl")
	case "top":
		data.Data = h.ReadFromCommand("top", "-b", "-n 1")
	}
	data.CurrentLocalMenuPos = page
	data.CurrentPage = page
	app.RenderTemplate(w, data, "base", "base.tmpl", "log.tmpl", "flash.tmpl")
}

type sevicesPageCtx struct {
	*PageCtx
	Services map[string]string
}

func servicesPageHangler(w http.ResponseWriter, r *http.Request) {
	ctx := &sevicesPageCtx{PageCtx: newNetPageCtx(w, r)}
	ctx.Services = make(map[string]string)
	lines := strings.Split(h.ReadFromCommand("service", "--status-all"), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 3 {
			ctx.Services[fields[3]] = fields[1]
		}
	}

	ctx.CurrentLocalMenuPos = "services"
	ctx.CurrentPage = "services"
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "services.tmpl", "flash.tmpl")

}

func serviceActionPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	service, ok := vars["service"]
	if !ok || service == "" {
		serviceActionPageHandler(w, r)
		return
	}
	action, ok := vars["action"]
	if !ok || action == "" {
		serviceActionPageHandler(w, r)
		return
	}
	l.Info("process serviceActionPageHandler %s %s", service, action)
	result := h.ReadFromCommand("sudo", "service", service, action)
	l.Info("process serviceActionPageHandler %s %s res=%s", service, action, result)
	session := app.GetSessionStore(w, r)
	session.Session.AddFlash(result)
	session.Save(w, r)
	http.Redirect(w, r, "/process/services", http.StatusFound)
}
