package pmain

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	h "k.prv/rpimon/helpers"
	"k.prv/rpimon/modules/monitor"
	"net/http"
	"runtime"
	"strings"
)

var Module = &context.Module{
	Name:          "system",
	Title:         "System",
	Description:   "",
	Init:          initModule,
	GetMenu:       getMenu,
	Internal:      true,
	AllPrivilages: []context.Privilege{context.Privilege{"admin", "system administrator"}},
}

// CreateRoutes for /main
func initModule(parentRoute *mux.Route) bool {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/",
		context.HandleWithContextSec(systemPageHandler, "System", "admin")).Name(
		"main-system")
	subRouter.HandleFunc("/serv/status",
		app.VerifyPermission(statusServHandler, "admin")).Name(
		"main-serv-status")
	return true
}

func getMenu(ctx *context.BasePageContext) (parentId string, menu *context.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		return "", nil
	}
	menu = context.NewMenuItem("System", "").SetIcon("glyphicon glyphicon-wrench").SetID("system")
	menu.AddChild(
		app.NewMenuItemFromRoute("Live view", "main-system").SetID("system-live").SetIcon("glyphicon glyphicon-dashboard"))

	return "", menu
}

type pageSystemCtx struct {
	*context.BasePageContext
	Warnings          *monitor.WarningsStruct
	MaxAcceptableLoad int
}

func systemPageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BasePageContext) {
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
