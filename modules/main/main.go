package pmain

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	"k.prv/rpimon/modules/monitor"
	"k.prv/rpimon/modules/mpd"
	"net/http"
	"runtime"
)

// Module information
var Module = &app.Module{
	Name:          "main",
	Title:         "Main",
	Description:   "Main",
	AllPrivilages: nil,
	Init:          initModule,
}

// CreateRoutes for /main
func initModule(parentRoute *mux.Route) bool {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.Context(mainPageHandler, "Main")).Name("main-index")
	subRouter.HandleFunc("/serv/alerts",
		app.VerifyPermission(alertsServHandler, "admin")).Name(
		"main-serv-alerts")
	return true
}

type pageCtx struct {
	*app.BaseCtx
	Uptime            *monitor.UptimeInfoStruct
	Load              *monitor.LoadInfoStruct
	CPUUsage          *monitor.CPUUsageInfoStruct
	CPUInfo           *monitor.CPUInfoStruct
	MemInfo           *monitor.MemInfo
	Filesystems       *monitor.FilesystemsStruct
	Interfaces        *monitor.InterfacesStruct
	MpdStatus         map[string]string
	Warnings          *app.WarningsStruct
	MaxAcceptableLoad int
	LoadTrucated      float64
	HostsStatus       map[string]bool
}

func mainPageHandler(r *http.Request, bctx *app.BaseCtx) {
	ctx := &pageCtx{BaseCtx: bctx}
	ctx.SetMenuActive("main")
	ctx.Warnings = app.GetWarnings()
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
	ctx.HostsStatus = monitor.GetSimpleHostStatus()
	ctx.RenderStd(ctx, "main/index.tmpl")
}

func alertsServHandler(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	data["warnings"] = app.GetWarnings()
	encoded, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(encoded)
}
