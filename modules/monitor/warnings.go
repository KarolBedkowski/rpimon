// Package monitor - system monitoring
package monitor

import (
	"k.prv/rpimon/app"
	"k.prv/rpimon/cfg"
	h "k.prv/rpimon/helpers"
	//	l "k.prv/rpimon/logging"
	"strconv"
	"strings"
)

// WARNINGS

const warningsCacheTTL = 5

var warningsCache = h.NewSimpleCache(warningsCacheTTL)

// GetWarnings return current warnings to show
func getWarnings() *app.WarningsStruct {
	result := warningsCache.Get(func() h.Value {
		cfg.Configuration.RLock()
		conf := cfg.Configuration.Monitor
		cfg.Configuration.RUnlock()
		warnings := &app.WarningsStruct{}
		// high load
		loadInfo := GetLoadInfo()
		if lastLoadInfo != nil {
			if conf.LoadError > 0 && loadInfo.Load5 > conf.LoadError {
				warnings.Errors = append(warnings.Errors, "Critical system Load")
			} else if conf.LoadWarning > 0 && loadInfo.Load5 > conf.LoadWarning {
				warnings.Warnings = append(warnings.Warnings, "High system Load")
			}
		}
		// low mem
		memInfo := GetMemoryInfo()
		if memInfo != nil {
			if conf.RAMUsageWarning > 0 && memInfo.UsedPerc > conf.RAMUsageWarning {
				if conf.SwapUsageWarning > 0 && memInfo.SwapTotal > 0 && memInfo.SwapFreePerc < 100-conf.SwapUsageWarning {
					warnings.Errors = append(warnings.Errors, "CRITICAL RAM/SWAP ussage")
				} else {
					warnings.Warnings = append(warnings.Warnings, "High memory ussage")
				}
			} else if conf.SwapUsageWarning > 0 && memInfo.SwapTotal > 0 && lastMemInfo.SwapFreePerc < 100-conf.SwapUsageWarning {
				warnings.Warnings = append(warnings.Warnings, "High SWAP ussage")
			}
		}
		// filesystems
		for _, fsinfo := range *GetFilesystemsInfo() {
			if fsinfo.Size == "0" {
				continue
			}
			if conf.DefaultFSUsageError > 0 && fsinfo.FreePerc < 100-conf.DefaultFSUsageError {
				warnings.Errors = append(warnings.Errors, "Low free space on "+fsinfo.Name)
			} else if conf.DefaultFSUsageWarning > 0 && fsinfo.FreePerc < 100-conf.DefaultFSUsageWarning {
				warnings.Warnings = append(warnings.Warnings, "Low free space on "+fsinfo.Name)
			}
		}
		// cpu temp
		cputemp := GetCPUInfo().Temp
		if conf.CPUTempError > 0 && cputemp > conf.CPUTempError {
			warnings.Errors = append(warnings.Errors, "Critical CPU temperature")
		} else if conf.CPUTempWarning > 0 && cputemp > conf.CPUTempWarning {
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
		for _, serv := range conf.MonitoredServices {
			//l.Debug("checking %v -> %v", port, comment)
			if checkIsServiceConnected(serv.Port) {
				warnings.Infos = append(warnings.Infos, serv.Name)
			}
		}
		/* test
		warnings.Warnings = append(warnings.Warnings, "Warn1", "Warn2")
		warnings.Errors = append(warnings.Errors, "Err1", "Err2")
		warnings.Infos = append(warnings.Infos, "Info1", "Info2")
		*/

		// Avalable hosts

		for _, host := range GetHostsStatus() {
			if !host.Available {
				switch host.MonitoredHost.Alarm {
				case 1: // info
					warnings.Infos = append(warnings.Infos, "Host "+host.MonitoredHost.Name+" unavailable")
				case 2: // warn
					warnings.Warnings = append(warnings.Warnings, "Host "+host.MonitoredHost.Name+" unavailable")
				case 3: // error
					warnings.Errors = append(warnings.Errors, "Host "+host.MonitoredHost.Name+" unavailable")
				}
			} else {
				switch host.MonitoredHost.Alarm {
				case 11: // info
					warnings.Infos = append(warnings.Infos, "Host "+host.MonitoredHost.Name+" available")
				case 12: // warn
					warnings.Warnings = append(warnings.Warnings, "Host "+host.MonitoredHost.Name+" available")
				case 13: // error
					warnings.Errors = append(warnings.Errors, "Host "+host.MonitoredHost.Name+" available")
				}
			}
		}

		return warnings

	}).(*app.WarningsStruct)
	return result
}

var netstatCache = h.NewSimpleCache(warningsCacheTTL)

func checkIsServiceConnected(port uint32) (result bool) {
	result = false
	out := netstatCache.Get(func() h.Value {
		return string(h.ReadCommand("netstat", "-pn", "-tu"))
	}).(string)
	if out == "" {
		return
	}
	lookingFor := ":" + strconv.Itoa(int(port))
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
