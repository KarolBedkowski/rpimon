package pmain

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	"k.prv/rpimon/modules/mpd"
	"k.prv/rpimon/monitor"
	"net/http"
	"runtime"
)

// CreateRoutes for /main
func CreateRoutes(parentRoute *mux.Route) {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", context.HandleWithContext(mainPageHandler, "Main")).Name("main-index")
	subRouter.HandleFunc("/serv/alerts",
		app.VerifyPermission(alertsServHandler, "admin")).Name(
		"main-serv-alerts")
}

type pageCtx struct {
	*context.BasePageContext
	Uptime            *monitor.UptimeInfoStruct
	Load              *monitor.LoadInfoStruct
	CPUUsage          *monitor.CPUUsageInfoStruct
	CPUInfo           *monitor.CPUInfoStruct
	MemInfo           *monitor.MemInfo
	Filesystems       *monitor.FilesystemsStruct
	Interfaces        *monitor.InterfacesStruct
	MpdStatus         map[string]string
	Warnings          *monitor.WarningsStruct
	MaxAcceptableLoad int
	LoadTrucated      float64
}

func mainPageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BasePageContext) {
	ctx := &pageCtx{BasePageContext: bctx}
	ctx.SetMenuActive("main")
	ctx.Warnings = monitor.GetWarnings()
	ctx.Uptime = monitor.GetUptimeInfo()
	ctx.CPUUsage = monitor.GetCPUUsageInfo()
	ctx.CPUInfo = monitor.GetCPUInfo()
	ctx.MemInfo = monitor.GetMemoryInfo()
	ctx.Load = monitor.GetLoadInfo()
	ctx.Interfaces = monitor.GetInterfacesInfo()
	ctx.Filesystems = monitor.GetFilesystemsInfo()
	ctx.MaxAcceptableLoad = runtime.NumCPU() * 2
	if ctx.Load.Load1 > float64(ctx.MaxAcceptableLoad) {
		ctx.LoadTrucated = float64(ctx.MaxAcceptableLoad)
	} else {
		ctx.LoadTrucated = ctx.Load.Load1
	}
	if mpdStatus, err := mpd.GetShortStatus(); err == nil {
		ctx.MpdStatus = mpdStatus
	}
	app.RenderTemplateStd(w, ctx, "main/index.tmpl")
}

func alertsServHandler(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	data["warnings"] = monitor.GetWarnings()
	encoded, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(encoded)
}
