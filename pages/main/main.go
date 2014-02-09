package pmain

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"k.prv/rpimon/monitor"
	"k.prv/rpimon/pages/mpd"
	"net/http"
	"runtime"
	"strings"
)

// CreateRoutes for /main
func CreateRoutes(parentRoute *mux.Route) {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", mainPageHanler).Name("main-index")
	subRouter.HandleFunc("/system",
		app.VerifyPermission(systemPageHanler, "admin")).Name(
		"main-system")
	subRouter.HandleFunc("/info",
		app.VerifyPermission(infoHandler, "admin")).Name(
		"main-serv-info")
}

type pageCtx struct {
	*app.BasePageContext
	Uptime            *monitor.UptimeInfoStruct
	Load              *monitor.LoadInfoStruct
	CPUUsage          *monitor.CPUUsageInfoStruct
	CPUInfo           *monitor.CPUInfoStruct
	MemInfo           *monitor.MemInfo
	Filesystems       *monitor.FilesystemsStruct
	Interfaces        *monitor.InterfacesStruct
	MpdStatus         map[string]string
	Warnings          []string
	MaxAcceptableLoad int
	LoadTrucated      float64
}

func mainPageHanler(w http.ResponseWriter, r *http.Request) {
	ctx := &pageCtx{BasePageContext: app.NewBasePageContext(
		"Main", "main", w, r)}
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

type pageSystemCtx struct {
	*app.BasePageContext
	Warnings          []string
	MaxAcceptableLoad int
}

func systemPageHanler(w http.ResponseWriter, r *http.Request) {
	ctx := &pageSystemCtx{BasePageContext: app.NewBasePageContext(
		"System", "system", w, r),
		Warnings: monitor.GetWarnings()}
	ctx.SetMenuActive("system-live")
	ctx.MaxAcceptableLoad = runtime.NumCPU() * 2
	app.RenderTemplateStd(w, ctx, "main/system.tmpl")
}

var infoHandlerCache = h.NewSimpleCache(1)

func infoHandler(w http.ResponseWriter, r *http.Request) {
	data := infoHandlerCache.Get(func() h.Value {
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
