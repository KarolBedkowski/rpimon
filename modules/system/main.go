package pmain

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"k.prv/rpimon/monitor"
	"net/http"
	"runtime"
	"strings"
)

var Module = &app.Module{
	Name:          "system",
	Title:         "System",
	Description:   "",
	AllPrivilages: nil,
	Init:          initModule,
	GetMenu:       getMenu,
}

// CreateRoutes for /main
func initModule(parentRoute *mux.Route) bool {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/",
		app.HandleWithContextSec(systemPageHandler, "System", "admin")).Name(
		"main-system")
	subRouter.HandleFunc("/serv/status",
		app.VerifyPermission(statusServHandler, "admin")).Name(
		"main-serv-status")
	return true
}

func getMenu(ctx *app.BasePageContext) (parentId string, menu *app.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		return "", nil
	}
	menu = app.NewMenuItem("System", "").SetIcon("glyphicon glyphicon-wrench").SetID("system")
	menu.AddChild(
		app.NewMenuItemFromRoute("Live view", "main-system").SetID("system-live").SetIcon("glyphicon glyphicon-dashboard"))

	return "", menu
}

type pageSystemCtx struct {
	*app.BasePageContext
	Warnings          *monitor.WarningsStruct
	MaxAcceptableLoad int
}

func systemPageHandler(w http.ResponseWriter, r *http.Request, bctx *app.BasePageContext) {
	ctx := &pageSystemCtx{BasePageContext: bctx,
		Warnings: monitor.GetWarnings()}
	ctx.SetMenuActive("system-live")
	ctx.MaxAcceptableLoad = runtime.NumCPU() * 2
	app.RenderTemplateStd(w, ctx, "main/system.tmpl")
}

var statusServCache = h.NewSimpleCache(1)

func statusServHandler(w http.ResponseWriter, r *http.Request) {
	data := statusServCache.Get(func() h.Value {
		res := map[string]interface{}{"cpu": strings.Join(monitor.GetCPUHistory(), ","),
			"load":     strings.Join(monitor.GetLoadHistory(), ","),
			"mem":      strings.Join(monitor.GetMemoryHistory(), ","),
			"meminfo":  monitor.GetMemoryInfo(),
			"cpuusage": monitor.GetCPUUsageInfo(),
			"cpuinfo":  monitor.GetCPUInfo(),
			"loadinfo": monitor.GetLoadInfo(),
			"fs":       monitor.GetFilesystemsInfo(),
			"iface":    monitor.GetInterfacesInfo(),
			"uptime":   monitor.GetUptimeInfo(),
			"netusage": monitor.GetTotalNetHistory(),
		}
		encoded, _ := json.Marshal(res)
		return encoded
	}).([]byte)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}
