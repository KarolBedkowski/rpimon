package pmain

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	//	"k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"k.prv/rpimon/monitor"
	"net/http"
	"os/exec"
	"strings"
)

var subRouter *mux.Router

// CreateRoutes for /main
func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", mainPageHanler).Name("main-index")
	subRouter.HandleFunc("/system", systemPageHanler).Name("main-system")
	subRouter.HandleFunc("/info", infoHandler)
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
	Uptime      string
	Users       string
	Load        *monitor.LoadInfoStruct
	CPUUsage    *monitor.CPUUsageInfoStruct
	CPUInfo     *monitor.CPUInfoStruct
	MemInfo     *monitor.MemInfo
	Filesystems *monitor.FilesystemsStruct
	Interfaces  *monitor.InterfacesStruct
	Warnings    []string
}

func mainPageHanler(w http.ResponseWriter, r *http.Request) {
	ctx := &pageCtx{BasePageContext: app.NewBasePageContext("Main", "main", w, r)}
	fillWarnings(ctx)
	fillUptimeInfo(ctx)
	ctx.CPUUsage = monitor.GetCPUUsageInfo()
	ctx.CPUInfo = monitor.GetCPUInfo()
	ctx.MemInfo = monitor.GetMemoryInfo()
	ctx.Load = monitor.GetLoadInfo()
	ctx.Interfaces = monitor.GetInterfacesInfo()
	ctx.Filesystems = monitor.GetFilesystemsInfo()
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "main/index.tmpl", "flash.tmpl")
}

func systemPageHanler(w http.ResponseWriter, r *http.Request) {
	ctx := &pageCtx{BasePageContext: app.NewBasePageContext("System", "system", w, r)}
	fillUptimeInfo(ctx)
	ctx.Filesystems = monitor.GetFilesystemsInfo()
	ctx.Interfaces = monitor.GetInterfacesInfo()
	fillWarnings(ctx)
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "main/system.tmpl", "flash.tmpl")
}

func fillUptimeInfo(ctx *pageCtx) error {
	out, err := exec.Command("uptime").Output()
	if err != nil {
		l.Warn("fillUptimeInfo Error", err)
		return err
	}
	fields := strings.SplitN(string(out), ",", 3)
	ctx.Uptime = strings.Join(strings.Fields(fields[0])[2:], " ")
	ctx.Users = strings.Split(strings.Trim(fields[1], " "), " ")[0]
	return nil
}

func fillWarnings(ctx *pageCtx) error {
	if checkIsServiceConnected("8200") {
		ctx.Warnings = append(ctx.Warnings, "MiniDLNA Connected")
	}
	if checkIsServiceConnected("445") {
		ctx.Warnings = append(ctx.Warnings, "SAMBA Connected")
	}
	if checkIsServiceConnected("21") {
		ctx.Warnings = append(ctx.Warnings, "FTP Connected")
	}
	return nil
}

func checkIsServiceConnected(port string) (result bool) {
	result = false
	out, err := exec.Command("netstat", "-pn", "--inet").Output()
	if err != nil {
		l.Warn("checkIsServiceConnected Error", err)
		return
	}
	outstr := string(out)
	lookingFor := ":" + port + " "
	if !strings.Contains(outstr, lookingFor) {
		return false
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if !strings.HasSuffix(line, "ESTABLISHED") {
			continue
		}
		if strings.Contains(line, lookingFor) {
			return true
		}
	}
	return
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
		"iface":    monitor.GetInterfacesInfo()}
	json.NewEncoder(w).Encode(res)
}
