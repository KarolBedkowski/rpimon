package network

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"k.prv/rpimon/modules"
	"k.prv/rpimon/monitor"
	"net/http"
	"strings"
)

func GetModule() *modules.Module {
	return &modules.Module{
		Name:          "network",
		Title:         "Network",
		Description:   "Network",
		AllPrivilages: nil,
		Init:          InitModule,
		GetMenu:       getMenu,
		GetWarnings:   getWarnings,
	}
}

func InitModule(parentRoute *mux.Route, configFilename string, conf *app.AppConfiguration) bool {
	// todo register modules
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.HandleWithContext(mainPageHandler,
		"Network")).Name("m-net")
	subRouter.HandleFunc("/conf", app.HandleWithContext(confPageHandler,
		"Network - Configuration")).Name("m-net-conf")
	subRouter.HandleFunc("/iptables", app.HandleWithContext(iptablesPageHandler,
		"Network - Iptables")).Name("m-net-iptables")
	subRouter.HandleFunc("/netstat", app.HandleWithContext(netstatPageHandler,
		"Network - Netstat")).Name("m-net-netstat")
	subRouter.HandleFunc("/serv/info", app.HandleWithContext(statusServHandler, "")).Name("m-net-serv-info")
	subRouter.HandleFunc("/action", app.HandleWithContext(actionHandler, "")).Name("m-net-action").Methods("PUT")
	return true
}

func getMenu(ctx *app.BasePageContext) (parentId string, menu *app.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		return "", nil
	}

	menu = app.NewMenuItemFromRoute("Network", "m-net").SetIcon("glyphicon glyphicon-dashboard")
	menu.Submenu = []*app.MenuItem{
		app.NewMenuItemFromRoute("Status", "m-net").SetID("m-net-index"),
		app.NewMenuItemFromRoute("Configuration", "m-net-conf"),
		app.NewMenuItemFromRoute("IPTables", "m-net-iptables"),
		app.NewMenuItemFromRoute("Netstat", "m-net-netstat"),
	}
	return "", menu
}

func getWarnings() map[string][]string {
	return nil
}

type mainPageContext struct {
	*app.BasePageContext
	Interfaces *monitor.InterfacesStruct
}

func mainPageHandler(w http.ResponseWriter, r *http.Request, ctx *app.BasePageContext) {
	c := &mainPageContext{BasePageContext: ctx}
	c.SetMenuActive("m-net-index")
	c.Interfaces = monitor.GetInterfacesInfo()
	app.RenderTemplateStd(w, c, "net/status.tmpl")
}

func netstatPageHandler(w http.ResponseWriter, r *http.Request, ctx *app.BasePageContext) {
	page := r.FormValue("sec")
	if page == "" {
		page = "listen"
	}
	data := &app.SimpleDataPageCtx{BasePageContext: ctx}
	data.SetMenuActive("m-net-netstat")
	data.THead = []string{"Proto", "Recv-Q", "Send-Q", "Local Address", "Port", "Foreign Address", "Port", "State", "PID", "Program name"}
	data.Header1 = "Netstat"
	switch page {
	case "listen":
		data.Header2 = "Listen"
		data.TData, _ = netstat("sudo", "netstat", "-lpn", "-t", "-p")
	case "connections":
		data.Header2 = "Connections"
		data.TData, _ = netstat("sudo", "netstat", "-pn", "-t", "-u")
	case "all":
		data.Header2 = "all"
		data.TData, _ = netstat("sudo", "netstat", "-apn", "-t", "-u")
	}
	data.Tabs = []*app.MenuItem{
		app.NewMenuItemFromRoute("Listen", "m-net-netstat").AddQuery("?sec=listen").SetActve(page == "listen"),
		app.NewMenuItemFromRoute("Connections", "m-net-netstat").AddQuery("?sec=connections").SetActve(page == "connections"),
		app.NewMenuItemFromRoute("All", "m-net-netstat").AddQuery("?sec=all").SetActve(page == "all"),
	}
	app.RenderTemplateStd(w, data, "data.tmpl", "tabs.tmpl")
}

type confPageContext struct {
	*app.BasePageContext
	Current  string
	Data     string
	Commands *map[string][]string
}

var confCommands = map[string][]string{
	"Base": []string{
		"ifconfig",
		"route -n",
		"arp -n",
		"cat /etc/hosts",
		"cat /etc/resolv.conf",
		"iwconfig",
	},
	"ip": []string{
		"ip link",
		"ip addr",
		"ip addrlabel",
		"ip route",
		"ip rule",
		"ip neigh",
		"ip ntable",
		"ip tunnel",
		"ip tuntap",
		"ip maddr",
		"ip mroute",
		"ip mrule",
		"ip monitor",
		"ip xfrm",
		"ip netns",
		"ip l2tp",
		"ip tcp_metrics",
		"ip token",
	},
	"iw": []string{
		"iw phy",
		"iw dev",
		"iw wlan0 scan dump",
		"iw wlan0 station dump",
		"iw wlan0 survey dump",
		"iw wlan0 link",
	},
}

func confPageHandler(w http.ResponseWriter, r *http.Request, ctx *app.BasePageContext) {
	cmd := r.FormValue("cmd")
	if cmd == "" {
		cmd = confCommands["Base"][0]
	} else {
		if !h.CheckValueInDictOfList(confCommands, cmd) {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
	}
	cmdfields := strings.Fields(cmd)

	if r.FormValue("data") == "1" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(h.ReadCommand(cmdfields[0], cmdfields[1:]...)))
	} else {
		ctx := &confPageContext{BasePageContext: ctx}
		ctx.SetMenuActive("m-net-conf")
		ctx.Current = cmd
		ctx.Commands = &confCommands
		ctx.Data = h.ReadCommand(cmdfields[0], cmdfields[1:]...)
		app.RenderTemplateStd(w, ctx, "net/conf.tmpl")
	}
}

type iptablesPageContext struct {
	*app.BasePageContext
	Current string
	Data    string
	Tables  *[]string
}

var iptablesTables = []string{
	"filter",
	"nat",
	"mangle",
	"raw",
	"security",
}

func iptablesPageHandler(w http.ResponseWriter, r *http.Request, ctx *app.BasePageContext) {
	table := r.FormValue("table")
	if table == "" {
		table = iptablesTables[0]
	} else {
		if !h.CheckValueInStrList(iptablesTables, table) {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
	}
	data := h.ReadCommand("sudo", "iptables", "-L", "-vn", "-t", table)

	if r.FormValue("data") == "1" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(data))
	} else {
		ctx := &iptablesPageContext{BasePageContext: ctx}
		ctx.SetMenuActive("m-net-iptables")
		ctx.Current = table
		ctx.Tables = &iptablesTables
		ctx.Data = data
		app.RenderTemplateStd(w, ctx, "net/iptables.tmpl")
	}
}

func netstat(command string, args ...string) ([][]string, error) {
	result := make([][]string, 0)
	res := h.ReadCommand(command, args...)
	lines := strings.Split(res, "\n")
	if len(lines) < 2 {
		return result, nil
	}
	for _, line := range lines[2:] {
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "tcp ") && !strings.HasPrefix(line, "udp ") {
			continue
		}
		fields := strings.Fields(line[:80])
		//l.Debug("%v\n%#v, %#v", line, fields)
		state := ""
		if len(fields) == 6 {
			state = fields[5]
		}
		var pidcmd []string
		pidcmdfield := strings.TrimSpace(line[80:])
		if pidcmdfield == "-" {
			pidcmd = []string{"", "-"}
		} else {
			pidcmd = strings.Split(pidcmdfield, "/")
		}
		//l.Debug("%v, %#v, %#v", line, fields, pidcmd)
		laddressDiv := strings.Split(fields[3], ":")
		faddressDiv := strings.Split(fields[4], ":")
		result = append(result, []string{
			fields[0], fields[1], fields[2],
			laddressDiv[0], laddressDiv[1],
			faddressDiv[0], faddressDiv[1],
			state, pidcmd[0], pidcmd[1],
		})
	}
	return result, nil
}

var statusServCache = h.NewSimpleCache(1)

func statusServHandler(w http.ResponseWriter, r *http.Request, ctx *app.BasePageContext) {
	data := statusServCache.Get(func() h.Value {
		res := map[string]interface{}{
			"netusage": monitor.GetNetHistory(),
			"ifaces":   monitor.GetInterfacesInfo(),
		}
		encoded, _ := json.Marshal(res)
		return encoded
	}).([]byte)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}

func actionHandler(w http.ResponseWriter, r *http.Request, ctx *app.BasePageContext) {
	action := r.FormValue("action")
	iface := r.FormValue("iface")
	if action == "" || iface == "" {
		http.Error(w, "missing action and/or iface", http.StatusBadRequest)
		return
	}

	var result string
	switch action {
	case "dhclient":
		result = h.ReadCommand("sudo", "dhclient", iface)
	case "down":
		result = h.ReadCommand("sudo", "ifconfig", iface, "down")
	case "up":
		result = h.ReadCommand("sudo", "ifconfig", iface, "up")
	default:
		http.Error(w, "wrong action", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(result))
}
