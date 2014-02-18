// Package monitor - system monitoring
package monitor

import (
	h "k.prv/rpimon/helpers"
	"runtime"
	"strings"
)

// WARNINGS

type WarningsStruct struct {
	Warnings []string
	Errors   []string
	Infos    []string
}

const warningsCacheTTL = 5

var maxAcceptableLoad = float64(runtime.NumCPU() * 2)

var warningsCache = h.NewSimpleCache(warningsCacheTTL)

// GetWarnings return current warnings to show
func GetWarnings() *WarningsStruct {
	result := warningsCache.Get(func() h.Value {
		warnings := &WarningsStruct{}
		// high load
		if lastLoadInfo != nil {
			if lastLoadInfo.Load5 > maxAcceptableLoad*2 {
				warnings.Errors = append(warnings.Errors, "Critical system Load")
			} else if lastLoadInfo.Load5 > maxAcceptableLoad {
				warnings.Warnings = append(warnings.Warnings, "High system Load")
			}
		}
		// low mem
		if lastMemInfo != nil && lastMemInfo.UsedPerc > 90 {
			if lastMemInfo.SwapFreePerc < 25 {
				warnings.Errors = append(warnings.Errors, "CRITICAL memory ussage")
			} else {
				warnings.Warnings = append(warnings.Warnings, "High memory ussage")
			}
		}
		// filesystems
		for _, fsinfo := range *GetFilesystemsInfo() {
			if fsinfo.Size == "0" {
				continue
			}
			if fsinfo.FreePerc < 5 {
				warnings.Errors = append(warnings.Errors, "Low free space on "+fsinfo.Name)
			} else if fsinfo.FreePerc < 10 {
				warnings.Warnings = append(warnings.Warnings, "Low free space on "+fsinfo.Name)
			}
		}
		// cpu temp
		cputemp := GetCPUInfo().Temp
		if cputemp > 80 {
			warnings.Errors = append(warnings.Errors, "Critical CPU temperature")
		} else if cputemp > 60 {
			warnings.Warnings = append(warnings.Warnings, "High CPU temperature")
		}
		// Services
		if checkIsServiceConnected("8200") {
			warnings.Infos = append(warnings.Infos, "MiniDLNA Connected")
		}
		if checkIsServiceConnected("445") {
			warnings.Infos = append(warnings.Infos, "SAMBA Connected")
		}
		if checkIsServiceConnected("21") {
			warnings.Infos = append(warnings.Infos, "FTP Connected")
		}
		/* test
		warnings.Warnings = append(warnings.Warnings, "Warn1", "Warn2")
		warnings.Errors = append(warnings.Errors, "Err1", "Err2")
		warnings.Infos = append(warnings.Infos, "Info1", "Info2")
		*/
		return warnings

	}).(*WarningsStruct)
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
