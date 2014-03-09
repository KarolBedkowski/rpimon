// Package monitor - system monitoring
package monitor

import (
	"k.prv/rpimon/app/cfg"
	h "k.prv/rpimon/helpers"
	//	l "k.prv/rpimon/helpers/logging"
	"strings"
)

// WARNINGS

// WarningsStruct holds current warnings, errors and informations.
type WarningsStruct struct {
	Warnings []string
	Errors   []string
	Infos    []string
}

const warningsCacheTTL = 5

var warningsCache = h.NewSimpleCache(warningsCacheTTL)

// GetWarnings return current warnings to show
func GetWarnings() *WarningsStruct {
	result := warningsCache.Get(func() h.Value {
		conf := cfg.Configuration.Monitor
		warnings := &WarningsStruct{}
		// high load
		if lastLoadInfo != nil {
			if lastLoadInfo.Load5 > conf.LoadError {
				warnings.Errors = append(warnings.Errors, "Critical system Load")
			} else if lastLoadInfo.Load5 > conf.LoadWarning {
				warnings.Warnings = append(warnings.Warnings, "High system Load")
			}
		}
		// low mem
		if lastMemInfo != nil {
			if lastMemInfo.UsedPerc > conf.RAMUsageWarning {
				if lastMemInfo.SwapTotal > 0 && lastMemInfo.SwapFreePerc < 100-conf.SwapUsageWarning {
					warnings.Errors = append(warnings.Errors, "CRITICAL RAM/SWAP ussage")
				} else {
					warnings.Warnings = append(warnings.Warnings, "High memory ussage")
				}
			} else if lastMemInfo.SwapTotal > 0 && lastMemInfo.SwapFreePerc < 100-conf.SwapUsageWarning {
				warnings.Warnings = append(warnings.Warnings, "High SWAP ussage")
			}
		}
		// filesystems
		for _, fsinfo := range *GetFilesystemsInfo() {
			if fsinfo.Size == "0" {
				continue
			}
			if fsinfo.FreePerc < 100-conf.DefaultFSUsageError {
				warnings.Errors = append(warnings.Errors, "Low free space on "+fsinfo.Name)
			} else if fsinfo.FreePerc < 100-conf.DefaultFSUsageWarning {
				warnings.Warnings = append(warnings.Warnings, "Low free space on "+fsinfo.Name)
			}
		}
		// cpu temp
		cputemp := GetCPUInfo().Temp
		if cputemp > conf.CPUTempError {
			warnings.Errors = append(warnings.Errors, "Critical CPU temperature")
		} else if cputemp > conf.CPUTempWarning {
			warnings.Warnings = append(warnings.Warnings, "High CPU temperature")
		}
		// Services
		/*
			if checkIsServiceConnected("8200") {
				warnings.Infos = append(warnings.Infos, "MiniDLNA Connected")
			}
			if checkIsServiceConnected("445") {
				warnings.Infos = append(warnings.Infos, "SAMBA Connected")
			}
			if checkIsServiceConnected("21") {
				warnings.Infos = append(warnings.Infos, "FTP Connected")
			}
		*/
		for port, comment := range conf.MonitoredServices {
			//l.Debug("checking %v -> %v", port, comment)
			if checkIsServiceConnected(port) {
				warnings.Infos = append(warnings.Infos, comment)
			}
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
		return string(h.ReadCommand("netstat", "-pn", "-tu"))
	}).(string)
	if out == "" {
		return
	}
	lookingFor := ":" + port
	if !strings.Contains(out, lookingFor) {
		return false
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if !strings.HasPrefix(line, "tcp ") && !strings.HasPrefix(line, "udp ") {
			continue
		}
		fields := strings.Fields(line)
		if fields[5] != "ESTABLISHED" {
			continue
		}
		if strings.HasSuffix(fields[3], lookingFor) {
			return true
		}
	}
	return
}
