// Package monitor - system monitoring
package monitor

import (
	"bufio"
	"k.prv/rpimon/app/cfg"
	l "k.prv/rpimon/helpers/logging"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

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
	netHistoryMutex.RLock()
	defer netHistoryMutex.RUnlock()
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
	netHistoryMutex.RLock()
	defer netHistoryMutex.RUnlock()
	result := ifaceHistory{
		Input:  netTotalUsage.Input[:],
		Output: netTotalUsage.Output[:],
	}
	return result
}

func gatherNetworkUsage() {
	netHistoryMutex.Lock()
	defer netHistoryMutex.Unlock()
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		l.Warn("gatherNetworkUsage open proc file error: %s", err.Error())
		return
	}
	defer file.Close()
	ts := time.Now().Unix()
	reader := bufio.NewReader(file)
	var sumRecv, sumTrans uint64
	for idx := 0; ; idx++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if idx < 2 {
			continue
		}
		sep := strings.Index(line, ":")
		iface := strings.TrimSpace(line[:sep])
		if iface == "lo" {
			continue
		}
		fields := strings.Fields(line[sep+1:])
		recv, err := strconv.ParseUint(fields[0], 10, 64)
		trans, err := strconv.ParseUint(fields[8], 10, 64)
		sumRecv += recv
		sumTrans += trans

		if ihist, ok := netHistory[iface]; ok {
			if ihist.lastTS > 0 {
				tsdelta := ts - ihist.lastTS
				if len(ihist.Input) >= netHistoryLimit {
					ihist.Input = ihist.Input[1:]
					ihist.Output = ihist.Output[1:]
				}
				if tsdelta > 0 {
					ihist.Input = append(ihist.Input, (recv-ihist.lastInput)/uint64(tsdelta))
					ihist.Output = append(ihist.Output, (trans-ihist.lastOutput)/uint64(tsdelta))
				}
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
		if tsdelta > 0 {
			netTotalUsage.Input = append(netTotalUsage.Input, (sumRecv-netTotalUsage.lastInput)/uint64(tsdelta))
			netTotalUsage.Output = append(netTotalUsage.Output, (sumTrans-netTotalUsage.lastOutput)/uint64(tsdelta))
		}
	}
	netTotalUsage.lastTS = ts
	netTotalUsage.lastInput = sumRecv
	netTotalUsage.lastOutput = sumTrans
}

type (
	hostStatus struct {
		lastCheck int
		available bool
	}

	// Host holds information about one monitored host.
	Host struct {
		cfg.MonitoredHost
		Available bool
	}
)

var (
	lastHostStatus      map[string]hostStatus
	monitoredHostsMutex sync.RWMutex
)

// GetHostsStatus return current host status
func GetHostsStatus() []*Host {
	monitoredHostsMutex.RLock()
	defer monitoredHostsMutex.RUnlock()
	result := make([]*Host, 0)
	for _, host := range cfg.Configuration.Monitor.MonitoredHosts {
		status, ok := lastHostStatus[host.Name]
		if ok {
			result = append(result, &Host{host, status.available})
		} else {
			result = append(result, &Host{host, false})
		}
	}
	return result
}

// GetSimpleHostStatus return current hosts status as map name->status
func GetSimpleHostStatus() map[string]bool {
	monitoredHostsMutex.RLock()
	defer monitoredHostsMutex.RUnlock()
	result := make(map[string]bool, 0)
	for _, host := range cfg.Configuration.Monitor.MonitoredHosts {
		if status, ok := lastHostStatus[host.Name]; ok {
			result[host.Name] = status.available
		} else {
			result[host.Name] = false
		}
	}
	return result
}

func checkHosts() {
	hosts := make(map[string]hostStatus, 0)
	now := int(time.Now().Unix())
	for _, chost := range cfg.Configuration.Monitor.MonitoredHosts {
		status, ok := lastHostStatus[chost.Name]
		if ok && status.lastCheck+chost.Interval >= now {
			hosts[chost.Name] = status
		} else {
			l.Debug("Monitor.checkHosts: checking %v", chost)
			available := false
			switch chost.Method {
			case "tcp":
				//_, err := exec.Command("nping", "--tcp-connect", "-H", "-N", "-c", "1", "-v-4", "-p", port, addr).CombinedOutput()
				conn, err := net.DialTimeout("tcp", chost.Address, time.Duration(1)*time.Second)
				if err == nil {
					defer conn.Close()
					_, err = conn.Write([]byte("\n"))
				}
				available = err == nil
			case "http":
				res, err := http.Get(chost.Address)
				available = err == nil && res.StatusCode >= 200 && res.StatusCode < 400
			default:
				_, err := exec.Command("ping", "-c", "1", "-i", "1", chost.Address).CombinedOutput()
				available = err == nil
			}
			now = int(time.Now().Unix())
			hosts[chost.Name] = hostStatus{now, available}
		}
	}
	monitoredHostsMutex.Lock()
	defer monitoredHostsMutex.Unlock()
	lastHostStatus = hosts
}
