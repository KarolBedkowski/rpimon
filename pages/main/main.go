package users

import (
	"bufio"
	"github.com/gorilla/mux"
	"io"
	"k.prv/rpimon/app"
	"k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var subRouter *mux.Router

// CreateRoutes for /main
func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", mainPageHanler).Name("main-index")
}

type fsInfo struct {
	Name       string
	Size       string
	Used       string
	Available  string
	UserPerc   string
	MountPoint string
}

type interfaceInfo struct {
	Name    string
	Address string
}

type pageCtx struct {
	*app.BasePageContext
	Hostname        string
	Uname           string
	Uptime          string
	Users           string
	Load            string
	CPUSystem       int
	CPUUser         int
	CPUIdle         int
	CPUIowait       int
	CPUFreq         int
	CPUMinFreq      int
	CPUMaxFreq      int
	CPUGovernor     string
	CPUTemp         int
	Filesystems     []fsInfo
	MemTotal        int
	MemFree         int
	MemBuffers      int
	MemCached       int
	MemSwapTotal    int
	MemSwapFree     int
	MemUsedPerc     int
	MemBuffersPerc  int
	MemCachePerc    int
	MemSwapUsedPerc int
	Interfaces      []interfaceInfo
	Warnings        []string
}

func newMainPageCtx(w http.ResponseWriter, r *http.Request) *pageCtx {
	return &pageCtx{BasePageContext: app.NewBasePageContext("Main", w, r)}
}

func mainPageHanler(w http.ResponseWriter, r *http.Request) {
	data := newMainPageCtx(w, r)
	fillUptimeInfo(data)
	fillCPUInfog(data)
	fillFSInfo(data)
	fillMemoryInfo(data)
	fillIfaceInfo(data)
	fillUnameInfo(data)
	fillWarnings(data)
	app.RenderTemplate(w, data, "base", "base.tmpl", "main/main.tmpl", "flash.tmpl")
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
	ctx.Load = strings.Split(fields[2], ":")[1]
	return nil
}

var lastUname string

func fillUnameInfo(ctx *pageCtx) error {
	if lastUname == "" {
		out, err := exec.Command("uname", "-mrsv").Output()
		if err != nil {
			l.Warn("fillUnameInfo Error", err)
			return err
		}
		lastUname = strings.Trim(string(out), " \n")
	}
	ctx.Uname = lastUname
	return nil
}

func fillCPUInfog(ctx *pageCtx) error {
	file, err := os.Open("/proc/stat")
	if err != nil {
		l.Warn("fillCPUInfog Error", err)
		return err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	line, _ := reader.ReadString('\n')
	line = strings.Replace(line, "  ", " ", -1)
	fields := strings.Fields(line)
	user, _ := strconv.Atoi(fields[1])
	user2, _ := strconv.Atoi(fields[2])
	user += user2
	system, _ := strconv.Atoi(fields[3])
	idle, _ := strconv.Atoi(fields[4])
	iowait, _ := strconv.Atoi(fields[5])
	usage := user + system + idle + iowait

	ctx.CPUSystem = int(100 * system / usage)
	ctx.CPUUser = int(100 * user / usage)
	ctx.CPUIdle = int(100 * idle / usage)
	ctx.CPUIowait = int(100 * iowait / usage)

	ctx.CPUFreq = helpers.ReadIntFromFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq") / 1000
	ctx.CPUMinFreq = helpers.ReadIntFromFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_min_freq") / 1000
	ctx.CPUMaxFreq = helpers.ReadIntFromFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_max_freq") / 1000
	ctx.CPUGovernor, _ = helpers.ReadLineFromFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor")

	ctx.CPUTemp = helpers.ReadIntFromFile("/sys/class/thermal/thermal_zone0/temp") / 1000

	return nil
}

func fillFSInfo(ctx *pageCtx) error {
	out, err := exec.Command("df", "-h", "-l", "-x", "tmpfs", "-x", "devtmpfs", "-x", "rootfs").Output()
	if err != nil {
		l.Warn("fillFSInfo Error", err)
		return err
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines[1:] {
		if len(line) == 0 {
			break
		}
		fields := strings.Fields(line)
		fsinfo := fsInfo{fields[0], fields[1], fields[2], fields[3], fields[4], fields[5]}
		ctx.Filesystems = append(ctx.Filesystems, fsinfo)
	}
	return nil
}

func fillMemoryInfo(ctx *pageCtx) error {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		l.Warn("fillCPUInfog Error", err)
		return err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		switch {
		case strings.HasPrefix(line, "MemTotal:"):
			ctx.MemTotal = getIntValueFromKeyVal(line)
		case strings.HasPrefix(line, "MemFree:"):
			ctx.MemFree = getIntValueFromKeyVal(line)
		case strings.HasPrefix(line, "Buffers:"):
			ctx.MemBuffers = getIntValueFromKeyVal(line)
		case strings.HasPrefix(line, "Cached:"):
			ctx.MemCached = getIntValueFromKeyVal(line)
		case strings.HasPrefix(line, "SwapFree:"):
			ctx.MemSwapFree = getIntValueFromKeyVal(line)
		case strings.HasPrefix(line, "SwapTotal:"):
			ctx.MemSwapTotal = getIntValueFromKeyVal(line)
		}
	}
	if ctx.MemTotal > 0 {
		ctx.MemUsedPerc = int(100 - 100.0*(ctx.MemFree+ctx.MemBuffers+ctx.MemCached)/ctx.MemTotal)
		ctx.MemBuffersPerc = int(100.0 * ctx.MemBuffers / ctx.MemTotal)
		ctx.MemCachePerc = int(100.0 * ctx.MemCached / ctx.MemTotal)
	}
	if ctx.MemSwapTotal > 0 {
		ctx.MemSwapUsedPerc = int(100 - 100.0*ctx.MemSwapFree/ctx.MemSwapTotal)
	}
	return nil
}

func fillIfaceInfo(ctx *pageCtx) error {
	out, err := exec.Command("/sbin/ip", "addr").Output()
	if err != nil {
		l.Warn("fillFSInfo Error", err)
		return err
	}
	lines := strings.Split(string(out), "\n")
	iface := ""
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if line[0] != ' ' {
			if iface != "" && iface != "lo" {
				ctx.Interfaces = append(ctx.Interfaces, interfaceInfo{iface, "-"})
			}
			iface = strings.Trim(strings.Fields(line)[1], " :")
		} else if strings.HasPrefix(line, "    inet") {
			if iface != "lo" {
				fields := strings.Fields(line)
				ctx.Interfaces = append(ctx.Interfaces,
					interfaceInfo{iface, fields[1]})
			}
			iface = ""
		}
	}
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

func getIntValueFromKeyVal(line string) int {
	fields := strings.Fields(line)
	res, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0
	}
	return res
}

func checkIsServiceConnected(port string) (result bool) {
	result = false
	out, err := exec.Command("netstat", "-pn", "--inet").Output()
	if err != nil {
		l.Warn("checkIsServiceConnected Error", err)
		return
	}
	lines := strings.Split(string(out), "\n")
	lookingFor := ":" + port + " "
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
