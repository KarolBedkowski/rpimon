package process

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	"k.prv/rpimon/app/session"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"strings"
)

// Module information
var Module = &context.Module{
	Name:          "system-process",
	Title:         "Process",
	Description:   "",
	AllPrivilages: nil,
	Init:          initModule,
	GetMenu:       getMenu,
}

// CreateRoutes for /process
func initModule(parentRoute *mux.Route) bool {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(psaxlPageHandler, "admin")).Name("process-index")
	subRouter.HandleFunc("/services", app.VerifyPermission(servicesPageHangler, "admin")).Name("process-services")
	subRouter.HandleFunc("/services/action", app.VerifyPermission(serviceActionPageHandler, "admin")).Name("process-services-action")
	subRouter.HandleFunc("/psaxl", app.VerifyPermission(psaxlPageHandler, "admin")).Name("process-psaxl")
	subRouter.HandleFunc("/top", app.VerifyPermission(topPageHandler, "admin")).Name("process-top")
	subRouter.HandleFunc("/process/action", app.VerifyPermission(processActionHandler, "admin")).Name(
		"process-action")
	return true
}

func getMenu(ctx *context.BasePageContext) (parentID string, menu *context.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		return "", nil
	}
	menu = app.NewMenuItemFromRoute("Process", "process-index").SetID("process").SetIcon("glyphicon glyphicon-cog")
	menu.AddChild(app.NewMenuItemFromRoute("PS AXL", "process-psaxl").SetID("psaxl"),
		app.NewMenuItemFromRoute("TOP", "process-top").SetID("top"),
		app.NewMenuItemFromRoute("Services", "process-services").SetID("services"),
	)
	return "system", menu
}

type sevicesPageCtx struct {
	*context.SimpleDataPageCtx
	Services map[string]string
}

func servicesPageHangler(w http.ResponseWriter, r *http.Request) {
	ctx := &sevicesPageCtx{SimpleDataPageCtx: context.NewSimpleDataPageCtx(
		w, r, "Process")}
	ctx.Services = make(map[string]string)
	lines := strings.Split(h.ReadCommand("service", "--status-all"), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 3 {
			ctx.Services[fields[3]] = fields[1]
		}
	}

	ctx.SetMenuActive("services")
	ctx.Header1 = "Services"
	app.RenderTemplateStd(w, ctx, "system/process/services.tmpl")

}

func serviceActionPageHandler(w http.ResponseWriter, r *http.Request) {
	service := r.FormValue("service")
	action := r.FormValue("action")
	if service == "" || action == "" {
		http.Error(w, "invalid request; missing service and/or action", http.StatusBadRequest)
		return
	}
	l.Info("process serviceActionPageHandler %s %s", service, action)
	result := h.ReadCommand("sudo", "service", service, action)
	s := session.GetSessionStore(w, r)
	s.AddFlash(result, "info")
	s.Save(r, w)
	http.Redirect(w, r, app.GetNamedURL("process-services"), http.StatusFound)
}

func psaxlPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &sevicesPageCtx{
		SimpleDataPageCtx: context.NewSimpleDataPageCtx(w, r, "Process"),
	}
	ctx.SetMenuActive("psaxl")
	ctx.Header1 = "Process"
	ctx.Header2 = "psaxl"

	lines := h.ReadCommand("ps", "axlww")
	var columns = 0
	for idx, line := range strings.Split(lines, "\n") {
		if line != "" {
			fields := strings.Fields(line)
			if idx == 0 {
				ctx.THead = fields
				columns = len(fields) - 1
			} else {
				if len(fields) > columns {
					cmd := strings.Join(fields[columns:], " ")
					fields = append(fields[:columns], cmd)
				}
				ctx.TData = append(ctx.TData, fields)
			}
		}
	}

	app.RenderTemplateStd(w, ctx, "system/process/psaxl.tmpl")
}

func topPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &sevicesPageCtx{
		SimpleDataPageCtx: context.NewSimpleDataPageCtx(w, r, "Process"),
	}
	ctx.SetMenuActive("top")
	ctx.Header1 = "Process"
	ctx.Header2 = "top"

	lines := strings.Split(h.ReadCommand("top", "-b", "-n", "1", "-w", "256"), "\n")

	// find header length
	var headerLen = 0
	for idx, line := range lines {
		if line == "" {
			headerLen = idx
			break
		}
	}
	ctx.Data = strings.Join(lines[:headerLen], "\n")

	var columns = 0
	for idx, line := range lines[headerLen+1:] {
		if line != "" {
			fields := strings.Fields(line)
			if idx == 0 {
				ctx.THead = fields
				columns = len(fields) - 1
			} else {
				if len(fields) > columns {
					cmd := strings.Join(fields[columns:], " ")
					fields = append(fields[:columns], cmd)
				}
				ctx.TData = append(ctx.TData, fields)
			}
		}
	}

	app.RenderTemplateStd(w, ctx, "system/process/top.tmpl")
}

func processActionHandler(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("a")
	pid := r.FormValue("pid")
	if action == "" || pid == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	back := r.FormValue("back")
	var result string

	switch action {
	case "kill":
		result = h.ReadCommand("sudo", "kill", "-9", pid)
	case "stop":
		result = h.ReadCommand("sudo", "kill", pid)
	default:
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}
	s := session.GetSessionStore(w, r)
	if result == "" {
		s.AddFlash("Process killed", "success")
	} else {
		s.AddFlash(result, "error")
	}
	s.Save(r, w)
	if back == "" {
		back = app.GetNamedURL("process-index")
	}
	http.Redirect(w, r, back, http.StatusFound)
}
