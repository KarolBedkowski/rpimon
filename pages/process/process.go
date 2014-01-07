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

// CreateRoutes for /process
func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(mainPageHandler, "admin")).Name("process-index")
	subRouter.HandleFunc("/services", app.VerifyPermission(servicesPageHangler, "admin")).Name("process-services")
	subRouter.HandleFunc("/services/{service}/{action}", app.VerifyPermission(serviceActionPageHandler, "admin"))
	subRouter.HandleFunc("/{page}", app.VerifyPermission(mainPageHandler, "admin")).Name("process-page")
}

var localMenu []*app.MenuItem

func createLocalMenu() []*app.MenuItem {
	if localMenu == nil {
		localMenu = []*app.MenuItem{app.NewMenuItemFromRoute("PS AXL", "process-page", "page", "psaxl").SetID("psaxl"),
			app.NewMenuItemFromRoute("TOP", "process-page", "page", "top").SetID("top"),
			app.NewMenuItemFromRoute("Services", "process-page", "page", "services").SetID("services")}
	}
	return localMenu
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.NewSimpleDataPageCtx(w, r, "Process", "process", "", createLocalMenu())
	vars := mux.Vars(r)
	page, ok := vars["page"]
	if !ok {
		page = "psaxl"
	}
	switch page {
	case "psaxl":
		data.Data = h.ReadFromCommand("ps", "axlww")
	case "top":
		data.Data = h.ReadFromCommand("top", "-b", "-n", "1", "-w", "1024")
	}
	data.CurrentLocalMenuPos = page
	data.CurrentPage = page
	app.RenderTemplate(w, data, "base", "base.tmpl", "log.tmpl", "flash.tmpl")
}

type sevicesPageCtx struct {
	*app.SimpleDataPageCtx
	Services map[string]string
}

func servicesPageHangler(w http.ResponseWriter, r *http.Request) {
	ctx := &sevicesPageCtx{SimpleDataPageCtx: app.NewSimpleDataPageCtx(
		w, r, "Process", "process", "", createLocalMenu())}
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
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "services.tmpl", "flash.tmpl", "pager.tmpl")

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
	session.AddFlash(result)
	session.Save(r, w)
	http.Redirect(w, r, "/process/services", http.StatusFound)
}
