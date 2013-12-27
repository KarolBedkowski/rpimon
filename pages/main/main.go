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

var subRouter *mux.Router

// CreateRoutes for /main
func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", mainPageHanler).Name("main-index")
	subRouter.HandleFunc("/system", app.VerifyPermission(systemPageHanler, "admin")).Name("main-system")
	subRouter.HandleFunc("/info", app.VerifyPermission(infoHandler, "admin"))
}

type fsInfo struct {
	Name       string
	Size       string
	Used       string
	Available  string
	UsedPerc   int
	MountPoint string
	FreePerc   int
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
	Warnings          []string
	MaxAcceptableLoad int
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
	ctx.MaxAcceptableLoad = runtime.NumCPU()
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "main/index.tmpl", "flash.tmpl")
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
	ctx.MaxAcceptableLoad = runtime.NumCPU()
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "main/system.tmpl", "flash.tmpl")
}


func infoHandler(w http.ResponseWriter, r *http.Request) {
	res := map[string]interface{}{"cpu": strings.Join(monitor.CPUHistory, ","),
		"load":     strings.Join(monitor.LoadHistory, ","),
		"mem":      strings.Join(monitor.MemHistory, ","),
		"meminfo":  monitor.GetMemoryInfo(),
		"cpuusage": monitor.GetCPUUsageInfo(),
		"cpuinfo":  monitor.GetCPUInfo(),
		"loadinfo": monitor.GetLoadInfo(),
		"fs":       monitor.GetFilesystemsInfo(),
		"iface":    monitor.GetInterfacesInfo(),
		"uptime":   monitor.GetUptimeInfo()}
	json.NewEncoder(w).Encode(res)
}
