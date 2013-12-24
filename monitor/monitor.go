package monitor

import (
	"bufio"
	"io"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"os"
	"strconv"
	"strings"
	"time"
)

func Init(interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				update()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

var LoadHistory = make([]string, 0)
var CPUHistory = make([]string, 0)
var MemHistory = make([]string, 0)

const limit = 30

type CPUUsageInfoStruct struct {
	User   int
	Idle   int
	System int
	IoWait int
	Usage  int
}

var lastCPUUsage *CPUUsageInfoStruct

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

var lastMemInfo *MemInfo

type CPUInfoStruct struct {
	Freq int
	Temp int
}

var lastCPUInfo *CPUInfoStruct

type LoadInfoStruct struct {
	Load []string
}

var lastLoadInfo *LoadInfoStruct

func update() {
	if load, err := h.ReadLineFromFile("/proc/loadavg"); err == nil {
		if len(LoadHistory) > limit {
			LoadHistory = LoadHistory[1:]
		}
		loadVal := strings.Fields(load)
		lastLoadInfo = &LoadInfoStruct{loadVal}
		LoadHistory = append(LoadHistory, loadVal[0])
	}
	if lastCPUUsage = gatherCPUUsageInfo(); lastCPUUsage != nil {
		if len(CPUHistory) > limit {
			CPUHistory = CPUHistory[1:]
		}
		CPUHistory = append(CPUHistory, strconv.Itoa(lastCPUUsage.Usage))
	}
	if lastMemInfo = gatherMemoryInfo(); lastMemInfo != nil {
		if len(MemHistory) > limit {
			MemHistory = MemHistory[1:]
		}
		MemHistory = append(MemHistory, strconv.Itoa(lastMemInfo.UsedPerc))
	}
	lastCPUInfo = gatherCPUInfo()
}

var (
	cpuLastUser   int
	cpuLastNice   int
	cpuLastIdle   int
	cpuLastSystem int
	cpuLastIoWait int
	cpuLastAll    int
)

func GetCPUUsageInfo() *CPUUsageInfoStruct {
	if lastCPUUsage == nil {
		return &CPUUsageInfoStruct{}
	}
	return lastCPUUsage
}

func gatherCPUUsageInfo() *CPUUsageInfoStruct {
	cpuusage := &CPUUsageInfoStruct{}
	line, err := h.ReadLineFromFile("/proc/stat")
	if err != nil {
		l.Warn("fillCPUInfog Error", err)
		return cpuusage
	}
	fields := strings.Fields(line)
	cUser, _ := strconv.Atoi(fields[1])
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
	return cpuusage
}

func GetMemoryInfo() *MemInfo {
	if lastMemInfo == nil {
		return &MemInfo{}
	}
	return lastMemInfo
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

func GetCPUInfo() *CPUInfoStruct {
	if lastCPUInfo == nil {
		return &CPUInfoStruct{}
	}
	return lastCPUInfo
}

func gatherCPUInfo() *CPUInfoStruct {
	info := &CPUInfoStruct{}
	info.Freq = h.ReadIntFromFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq") / 1000
	info.Temp = h.ReadIntFromFile("/sys/class/thermal/thermal_zone0/temp") / 1000
	return info
}

func GetLoadInfo() *LoadInfoStruct {
	if lastLoadInfo == nil {
		return &LoadInfoStruct{}
	}
	return lastLoadInfo
}
