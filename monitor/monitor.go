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
	CPUUser   int
	CPUIdle   int
	CPUSystem int
	CPUIoWait int
	Usage     int
}

var lastCPUUsage *CPUUsageInfoStruct

type MemInfo struct {
	MemTotal        int
	MemFree         int
	MemFreeUser     int
	MemBuffers      int
	MemCached       int
	MemSwapTotal    int
	MemSwapFree     int
	MemUsedPerc     int
	MemBuffersPerc  int
	MemCachePerc    int
	MemFreePerc     int
	MemFreeUserPerc int
	MemSwapUsedPerc int
	MemSwapFreePerc int
}

var lastMemInfo *MemInfo

type CPUInfoStruct struct {
	CPUFreq int
	CPUTemp int
}

var lastCPUInfo *CPUInfoStruct

func update() {
	if load, err := h.ReadLineFromFile("/proc/loadavg"); err == nil {
		if len(LoadHistory) > limit {
			LoadHistory = LoadHistory[1:]
		}
		loadVal := strings.SplitN(load, " ", 2)
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
		MemHistory = append(MemHistory, strconv.Itoa(lastMemInfo.MemUsedPerc))
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
	cpuusage.CPUUser = int(100 * (cUser + cNice) / allDiff)
	cpuusage.CPUIdle = int(100 * cIdle / allDiff)
	cpuusage.CPUSystem = int(100 * cSystem / allDiff)
	cpuusage.CPUIoWait = int(100 * cIoWait / allDiff)
	cpuusage.Usage = 100 - cpuusage.CPUIdle
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
			meminfo.MemTotal = getIntValueFromKeyVal(line)
		case strings.HasPrefix(line, "MemFree:"):
			meminfo.MemFree = getIntValueFromKeyVal(line)
		case strings.HasPrefix(line, "Buffers:"):
			meminfo.MemBuffers = getIntValueFromKeyVal(line)
		case strings.HasPrefix(line, "Cached:"):
			meminfo.MemCached = getIntValueFromKeyVal(line)
		case strings.HasPrefix(line, "SwapFree:"):
			meminfo.MemSwapFree = getIntValueFromKeyVal(line)
		case strings.HasPrefix(line, "SwapTotal:"):
			meminfo.MemSwapTotal = getIntValueFromKeyVal(line)
		}
	}
	if meminfo.MemTotal > 0 {
		meminfo.MemUsedPerc = int(100 - 100.0*(meminfo.MemFree+meminfo.MemBuffers+meminfo.MemCached)/meminfo.MemTotal)
		meminfo.MemBuffersPerc = int(100.0 * meminfo.MemBuffers / meminfo.MemTotal)
		meminfo.MemCachePerc = int(100.0 * meminfo.MemCached / meminfo.MemTotal)
		meminfo.MemFreePerc = int(100 * meminfo.MemFree / meminfo.MemTotal)
		meminfo.MemFreeUserPerc = 100 - meminfo.MemUsedPerc
	}
	if meminfo.MemSwapTotal > 0 {
		meminfo.MemSwapFreePerc = int(100.0 * meminfo.MemSwapFree / meminfo.MemSwapTotal)
		meminfo.MemSwapUsedPerc = 100 - meminfo.MemSwapFreePerc
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
	info.CPUFreq = h.ReadIntFromFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq") / 1000
	info.CPUTemp = h.ReadIntFromFile("/sys/class/thermal/thermal_zone0/temp") / 1000
	return info
}
