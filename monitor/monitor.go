// Package monitor - system monitoring
package monitor

import (
	"bufio"
	"io"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Caches max TTL
const (
	historyLimit       = 30
	ifaceCacheTTL      = 5
	fsCacheTTL         = 10
	uptimeInfoCacheTTL = 2
	cpuInfoCacheTTL    = 5
	netHistoryLimit    = 30
)

// Init monitor, start background go routine
func Init(interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				gatherLoadInfo()
				gatherCPUUsageInfo()
				gatherMemoryInfo()
				gatherNetworkUsage()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

// CPUUsageInfoStruct - information about cpu usage
type CPUUsageInfoStruct struct {
	User   int
	Idle   int
	System int
	IoWait int
	Usage  int
}

var (
	cpuLastUser   int
	cpuLastNice   int
	cpuLastIdle   int
	cpuLastSystem int
	cpuLastIoWait int
	cpuLastAll    int
	lastCPUUsage  *CPUUsageInfoStruct
	cpuUsageMutex sync.RWMutex
	cpuHistory    = make([]string, 0)
)

// GetCPUUsageInfo - get last cpu usage information
func GetCPUUsageInfo() *CPUUsageInfoStruct {
	cpuUsageMutex.RLock()
	defer cpuUsageMutex.RUnlock()
	if lastCPUUsage == nil {
		return &CPUUsageInfoStruct{}
	}
	return lastCPUUsage
}

// GetCPUHistory - get cpu total usage information
func GetCPUHistory() []string {
	cpuUsageMutex.RLock()
	defer cpuUsageMutex.RUnlock()
	return []string(cpuHistory)
}

func gatherCPUUsageInfo() *CPUUsageInfoStruct {
	cpuusage := &CPUUsageInfoStruct{}
	line, err := h.ReadLineFromFile("/proc/stat")
	if err != nil {
		l.Warn("fillCPUInfog Error", err)
		return cpuusage
	}

	fields := strings.Fields(line)

	cpuUsageMutex.Lock()
	defer cpuUsageMutex.Unlock()

	cUser, err := strconv.Atoi(fields[1])
	cpuLastUser, cUser = cUser, cUser-cpuLastUser
	cNice, _ := strconv.Atoi(fields[2])
	cpuLastNice, cNice = cNice, cNice-cpuLastNice
	cSystem, _ := strconv.Atoi(fields[3])
	cpuLastSystem, cSystem = cSystem, cSystem-cpuLastSystem
	cIdle, _ := strconv.Atoi(fields[4])
	cpuLastIdle, cIdle = cIdle, cIdle-cpuLastIdle
	cIoWait, _ := strconv.Atoi(fields[5])
	cpuLastIoWait, cIoWait = cIoWait, cIoWait-cpuLastIoWait
	allDiff := cUser + cNice + cSystem + cIdle + cIoWait
	cpuusage.User = int(100 * (cUser + cNice) / allDiff)
	cpuusage.Idle = int(100 * cIdle / allDiff)
	cpuusage.System = int(100 * cSystem / allDiff)
	cpuusage.IoWait = int(100 * cIoWait / allDiff)
	cpuusage.Usage = 100 - cpuusage.Idle

	lastCPUUsage = cpuusage
	if len(cpuHistory) > historyLimit {
		cpuHistory = cpuHistory[1:]
	}
	cpuHistory = append(cpuHistory, strconv.Itoa(cpuusage.Usage))
	return cpuusage
}

// MemInfo - memory usage information
type MemInfo struct {
	Total        int
	Free         int
	FreeUser     int
	Buffers      int
	Cached       int
	SwapTotal    int
	SwapFree     int
	UsedPerc     int
	BuffersPerc  int
	CachePerc    int
	FreePerc     int
	FreeUserPerc int
	SwapUsedPerc int
	SwapFreePerc int
}

var (
	lastMemInfo     *MemInfo
	memoryInfoMutex sync.RWMutex
	memHistory      = make([]string, 0)
)

// GetMemoryInfo - get last memory usage
func GetMemoryInfo() *MemInfo {
	memoryInfoMutex.RLock()
	defer memoryInfoMutex.RUnlock()
	if lastMemInfo == nil {
		return &MemInfo{}
	}
	return lastMemInfo
}

// GetMemoryHistory get history of total memory usage
func GetMemoryHistory() []string {
	memoryInfoMutex.RLock()
	defer memoryInfoMutex.RUnlock()
	return []string(memHistory)
}

func gatherMemoryInfo() *MemInfo {
	meminfo := &MemInfo{}
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		l.Warn("fillCPUInfog Error", err)
		return meminfo
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
			meminfo.Total = getIntValueFromKeyVal(line)
		case strings.HasPrefix(line, "MemFree:"):
			meminfo.Free = getIntValueFromKeyVal(line)
		case strings.HasPrefix(line, "Buffers:"):
			meminfo.Buffers = getIntValueFromKeyVal(line)
		case strings.HasPrefix(line, "Cached:"):
			meminfo.Cached = getIntValueFromKeyVal(line)
		case strings.HasPrefix(line, "SwapFree:"):
			meminfo.SwapFree = getIntValueFromKeyVal(line)
		case strings.HasPrefix(line, "SwapTotal:"):
			meminfo.SwapTotal = getIntValueFromKeyVal(line)
		}
	}
	if meminfo.Total > 0 {
		meminfo.UsedPerc = int(100 - 100.0*(meminfo.Free+meminfo.Buffers+meminfo.Cached)/meminfo.Total)
		meminfo.BuffersPerc = int(100.0 * meminfo.Buffers / meminfo.Total)
		meminfo.CachePerc = int(100.0 * meminfo.Cached / meminfo.Total)
		meminfo.FreePerc = int(100 * meminfo.Free / meminfo.Total)
		meminfo.FreeUserPerc = 100 - meminfo.UsedPerc
	}
	if meminfo.SwapTotal > 0 {
		meminfo.SwapFreePerc = int(100.0 * meminfo.SwapFree / meminfo.SwapTotal)
		meminfo.SwapUsedPerc = 100 - meminfo.SwapFreePerc
	}

	memoryInfoMutex.Lock()
	defer memoryInfoMutex.Unlock()

	lastMemInfo = meminfo
	if len(memHistory) > historyLimit {
		memHistory = memHistory[1:]
	}
	memHistory = append(memHistory, strconv.Itoa(lastMemInfo.UsedPerc))
	return meminfo
}

func getIntValueFromKeyVal(line string) int {
	fields := strings.Fields(line)
	res, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0
	}
	return res
}

// CPUInfoStruct - information about frequency and temperature
type CPUInfoStruct struct {
	Freq int
	Temp int
}

var cpuInfoCache = h.NewSimpleCache(cpuInfoCacheTTL)

// GetCPUInfo get last cpu information
func GetCPUInfo() *CPUInfoStruct {
	result := cpuInfoCache.Get(func() h.Value {
		return gatherCPUInfo()
	})
	return result.(*CPUInfoStruct)
}

func gatherCPUInfo() *CPUInfoStruct {
	info := &CPUInfoStruct{}
	if val, err := h.ReadIntFromFile(app.Configuration.Monitor.CPUFreqFile); err == nil {
		info.Freq = val / 1000
	}
	if val, err := h.ReadIntFromFile(app.Configuration.Monitor.CPUTempFile); err == nil {
		info.Temp = val / 1000
	}
	return info
}

// LoadInfoStruct information about system load
type LoadInfoStruct struct {
	Load1  float64
	Load5  float64
	Load15 float64
}

var (
	lastLoadInfo *LoadInfoStruct
	loadMutex    sync.RWMutex
	loadHistory  = make([]string, 0)
)

// GetLoadHistory get history of system load
func GetLoadHistory() []string {
	loadMutex.RLock()
	defer loadMutex.RUnlock()
	return []string(loadHistory)
}

// GetLoadInfo get current load
func GetLoadInfo() *LoadInfoStruct {
	loadMutex.RLock()
	defer loadMutex.RUnlock()
	if lastLoadInfo == nil {
		return new(LoadInfoStruct)
	}
	return lastLoadInfo
}

func gatherLoadInfo() (err error) {
	if load, err := h.ReadLineFromFile("/proc/loadavg"); err == nil {
		loadMutex.Lock()
		defer loadMutex.Unlock()
		if len(loadHistory) > historyLimit {
			loadHistory = loadHistory[1:]
		}
		loadVal := strings.Fields(load)
		load1, _ := strconv.ParseFloat(loadVal[0], 10)
		load5, _ := strconv.ParseFloat(loadVal[1], 10)
		load15, _ := strconv.ParseFloat(loadVal[2], 10)
		lastLoadInfo = &LoadInfoStruct{load1, load5, load15}
		loadHistory = append(loadHistory, loadVal[0])
	}
	return
}

// InterfaceInfoStruct information about network interfaces
type InterfaceInfoStruct struct {
	Name     string
	Address  string
	Address6 string
	State    string
	Mac      string
	Kind     string
}

// InterfacesStruct informations about all interfaces
type InterfacesStruct []*InterfaceInfoStruct

var interfacesInfoCache = h.NewSimpleCache(ifaceCacheTTL)

func parseIPResult(input string) (result InterfacesStruct) {
	lines := strings.Split(input, "\n")
	var iface *InterfaceInfoStruct
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		fields := strings.Fields(line)
		if line[0] != ' ' {
			if iface != nil {
				if iface.Name != "lo" {
					result = append(result, iface)
				}
			}
			iface = new(InterfaceInfoStruct)
			iface.Name = strings.Trim(fields[1], " :")
			iface.State = fields[8]
		} else if strings.HasPrefix(line, "    inet ") {
			iface.Address = fields[1]
		} else if strings.HasPrefix(line, "    inet6 ") {
			iface.Address6 = fields[1]
		} else if strings.HasPrefix(line, "    link/") {
			iface.Mac = fields[1]
			iface.Kind = fields[0][5:]
		}
	}
	if iface != nil && iface.Name != "" && iface.Name != "lo" {
		result = append(result, iface)
	}
	return result
}

// GetInterfacesInfo get current info about network interfaces
func GetInterfacesInfo() *InterfacesStruct {
	result := interfacesInfoCache.Get(func() h.Value {
		ipres := h.ReadCommand("/sbin/ip", "addr")
		if ipres == "" {
			return nil
		}
		result := parseIPResult(ipres)
		return &result
	})
	return result.(*InterfacesStruct)
}

// FsInfoStruct information about filesystem mount & usage
type FsInfoStruct struct {
	Name       string
	Size       string
	Used       string
	Available  string
	UsedPerc   int
	MountPoint string
	FreePerc   int
}

// FilesystemsStruct list of FsInfoStruct
type FilesystemsStruct []FsInfoStruct

var fsInfoCache = h.NewSimpleCache(fsCacheTTL)

// GetFilesystemsInfo returns information about all filesystems
func GetFilesystemsInfo() *FilesystemsStruct {
	result := fsInfoCache.Get(func() h.Value {
		cmdout := h.ReadCommand("df", "-h", "-l", "-x", "tmpfs", "-x", "devtmpfs", "-x", "rootfs")
		if cmdout == "" {
			return nil
		}
		lines := strings.Split(cmdout, "\n")
		var result FilesystemsStruct
		for _, line := range lines[1:] {
			if len(line) == 0 {
				break
			}
			fields := strings.Fields(line)
			usedperc, _ := strconv.Atoi(strings.Trim(fields[4], "%"))
			result = append(result, FsInfoStruct{fields[0], fields[1], fields[2],
				fields[3], usedperc, fields[5], 100 - usedperc})
		}
		return &result
	})
	return result.(*FilesystemsStruct)
}

// UptimeInfoStruct information about uptime & users
type UptimeInfoStruct struct {
	Uptime string
	Users  string
}

var uptimeInfoCache = h.NewSimpleCache(uptimeInfoCacheTTL)

// GetUptimeInfo get current info about uptime & users
func GetUptimeInfo() *UptimeInfoStruct {
	result := uptimeInfoCache.Get(func() h.Value {
		cmdout := h.ReadCommand("uptime")
		if cmdout == "" {
			return nil
		}
		fields := strings.SplitN(cmdout, ",", 3)
		info := &UptimeInfoStruct{strings.Join(strings.Fields(fields[0])[2:], " "),
			strings.Split(strings.Trim(fields[1], " "), " ")[0]}
		return info
	})
	return result.(*UptimeInfoStruct)
}

type (
	// Total history as list of inputs and outputs
	ifaceHistory struct {
		lastTS     int64
		Input      []uint64
		Output     []uint64
		lastInput  uint64
		lastOutput uint64
	}
)

var (
	netHistoryMutex sync.RWMutex
	netHistory      = make(map[string]*ifaceHistory)
	netTotalUsage   ifaceHistory
)

// GetNetHistory return map interface->usage history
func GetNetHistory() map[string]ifaceHistory {
	loadMutex.RLock()
	defer loadMutex.RUnlock()
	result := make(map[string]ifaceHistory)
	for key, val := range netHistory {
		result[key] = ifaceHistory{
			Input:  val.Input[:],
			Output: val.Output[:],
		}
	}
	return result
}

// GetTotalNetHistory return ussage all interfaces
func GetTotalNetHistory() ifaceHistory {
	loadMutex.RLock()
	defer loadMutex.RUnlock()
	result := ifaceHistory{
		Input:  netTotalUsage.Input[:],
		Output: netTotalUsage.Output[:],
	}
	return result
}

func gatherNetworkUsage() {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		l.Warn("gatherNetworkUsage open proc file error: %s", err.Error())
		return
	}
	defer file.Close()
	ts := time.Now().Unix()
	reader := bufio.NewReader(file)
	netHistoryMutex.Lock()
	defer netHistoryMutex.Unlock()
	var sumRecv, sumTrans uint64
	for idx := 0; ; idx++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if idx < 2 {
			continue
		}
		fields := strings.Fields(line)
		iface := strings.Trim(fields[0], " :")
		if iface == "lo" {
			continue
		}

		recv, _ := strconv.ParseUint(fields[1], 10, 64)
		trans, _ := strconv.ParseUint(fields[9], 10, 64)
		sumRecv += recv
		sumTrans += trans

		if ihist, ok := netHistory[iface]; ok {
			if ihist.lastTS > 0 {
				tsdelta := ts - ihist.lastTS
				if len(ihist.Input) >= netHistoryLimit {
					ihist.Input = ihist.Input[1:]
					ihist.Output = ihist.Output[1:]
				}
				ihist.Input = append(ihist.Input, (recv-ihist.lastInput)/uint64(tsdelta))
				ihist.Output = append(ihist.Output, (trans-ihist.lastOutput)/uint64(tsdelta))
			}
			ihist.lastTS = ts
			ihist.lastInput = recv
			ihist.lastOutput = trans
		} else {
			ihist := ifaceHistory{
				lastTS:     ts,
				lastInput:  recv,
				lastOutput: trans,
				Input:      make([]uint64, 0),
				Output:     make([]uint64, 0),
			}
			netHistory[iface] = &ihist
		}
	}
	if netTotalUsage.lastTS > 0 {
		tsdelta := ts - netTotalUsage.lastTS
		if len(netTotalUsage.Input) >= netHistoryLimit {
			netTotalUsage.Input = netTotalUsage.Input[1:]
			netTotalUsage.Output = netTotalUsage.Output[1:]
		}
		netTotalUsage.Input = append(netTotalUsage.Input, (sumRecv-netTotalUsage.lastInput)/uint64(tsdelta))
		netTotalUsage.Output = append(netTotalUsage.Output, (sumTrans-netTotalUsage.lastOutput)/uint64(tsdelta))
	}
	netTotalUsage.lastTS = ts
	netTotalUsage.lastInput = sumRecv
	netTotalUsage.lastOutput = sumTrans
}
