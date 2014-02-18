// Package monitor - system monitoring
package monitor

import (
	h "k.prv/rpimon/helpers"
	"runtime"
	"strings"
)

// WARNINGS

var warningsCache = h.NewSimpleCache(warningsCacheTTL)
var maxAcceptableLoad = float64(runtime.NumCPU() * 2)

// GetWarnings return current warnings to show
func GetWarnings() []string {
	result := warningsCache.Get(func() h.Value {
		var warnings []string
		// high load
		if lastLoadInfo != nil && lastLoadInfo.Load5 > maxAcceptableLoad {
			warnings = append(warnings, "High system Load")
		}
		// low mem
		if lastMemInfo != nil && lastMemInfo.UsedPerc > 90 {
			if lastMemInfo.SwapFreePerc < 25 {
				warnings = append(warnings, "CRITICAL memory ussage")
			} else {
				warnings = append(warnings, "High memory ussage")
			}
		}
		// filesystems
		for _, fsinfo := range *GetFilesystemsInfo() {
			if fsinfo.FreePerc < 10 {
				warnings = append(warnings, "Low free space on "+fsinfo.Name)
			}
		}
		// cpu temp
		cputemp := GetCPUInfo().Temp
		if cputemp > 80 {
			warnings = append(warnings, "Critical CPU temperature")
		} else if cputemp > 60 {
			warnings = append(warnings, "High CPU temperature")
		}
		// Services
		if checkIsServiceConnected("8200") {
			warnings = append(warnings, "MiniDLNA Connected")
		}
		if checkIsServiceConnected("445") {
			warnings = append(warnings, "SAMBA Connected")
		}
		if checkIsServiceConnected("21") {
			warnings = append(warnings, "FTP Connected")
		}
		return warnings
	}).([]string)
	return result
}

var netstatCache = h.NewSimpleCache(warningsCacheTTL)

func checkIsServiceConnected(port string) (result bool) {
	result = false
	out := netstatCache.Get(func() h.Value {
		return string(h.ReadCommand("netstat", "-pn", "--inet"))
	}).(string)
	if out == "" {
		return
	}
	lookingFor := ":" + port + " "
	if !strings.Contains(out, lookingFor) {
		return false
	}
	lines := strings.Split(out, "\n")
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
