package users

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	//	l "k.prv/rpimon/helpers/logging"
	"k.prv/rpimon/monitor"
	"net/http"
	"strings"
)

// CreateRoutes for /net
func CreateRoutes(parentRoute *mux.Route) {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(mainPageHandler, "admin")).Name("net-index")
	subRouter.HandleFunc("/conf", app.VerifyPermission(confPageHandler, "admin")).Name("net-conf")
	subRouter.HandleFunc("/iptables", app.VerifyPermission(iptablesPageHandler, "admin")).Name("net-iptables")
	subRouter.HandleFunc("/serv/info", app.VerifyPermission(statusServHandler, "admin")).Name("net-serv-info")
	subRouter.HandleFunc("/action", app.VerifyPermission(actionHandler, "admin")).Name("net-action").Methods("PUT")
	subRouter.HandleFunc("/{page}", app.VerifyPermission(subPageHandler, "admin")).Name("net-page")
}

func buildLocalMenu() (localMenu []*app.MenuItem) {
	return []*app.MenuItem{
		app.NewMenuItemFromRoute("Status", "net-index").SetID("net-index"),
		app.NewMenuItemFromRoute("Configuration", "net-conf").SetID("conf"),
		app.NewMenuItemFromRoute("IPTables", "net-iptables").SetID("iptables"),
		app.NewMenuItemFromRoute("Netstat", "net-page", "page", "netstat").SetID("netstat"),
		app.NewMenuItemFromRoute("Conenctions", "net-page", "page", "connenctions").SetID("connenctions"),
		app.NewMenuItemFromRoute("Samba", "net-page", "page", "samba").SetID("samba"),
		app.NewMenuItemFromRoute("NFS", "net-page", "page", "nfs").SetID("nfs"),
	}
}

type mainPageContext struct {
	*app.BasePageContext
	Interfaces *monitor.InterfacesStruct
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &mainPageContext{BasePageContext: app.NewBasePageContext("Network", "net", w, r)}
	app.AttachSubmenu(ctx.BasePageContext, "net", buildLocalMenu())
	ctx.SetMenuActive("net-index")
	ctx.Interfaces = monitor.GetInterfacesInfo()
	app.RenderTemplateStd(w, ctx, "net/status.tmpl")
}

func subPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page, ok := vars["page"]
	if !ok {
		http.Redirect(w, r, app.GetNamedURL("net-index"), http.StatusFound)
		return
	}
	data := app.NewSimpleDataPageCtx(w, r, "Network", "net", page, buildLocalMenu())
	data.SetMenuActive(page)
	switch page {
	case "netstat":
		data.Header1 = "Netstat"
		data.Header2 = "Listen"
		data.THead = []string{"Proto", "Recv-Q", "Send-Q", "Local Address", "Port", "Foreign Address", "Port", "State", "PID", "Program name"}
		data.TData, _ = netstat("sudo", "netstat", "-lpn", "--inet", "--inet6")
	case "connenctions":
		data.Header1 = "Netstat"
		data.Header2 = "Connections"
		data.THead = []string{"Proto", "Recv-Q", "Send-Q", "Local Address", "Port", "Foreign Address", "Port", "State", "PID", "Program name"}
		data.TData, _ = netstat("sudo", "netstat", "-pn", "--inet", "--inet6")
	case "samba":
		data.Header1 = "Samba"
		data.Data = h.ReadCommand("sudo", "smbstatus")
	case "nfs":
		data.Header1 = "NFS Stat"
		data.Data = h.ReadCommand("nfsstat")
	}
	app.RenderTemplateStd(w, data, "data.tmpl")
}

type confPageContext struct {
	*app.BasePageContext
	Current  string
	Data     string
	Commands *[]string
}

var confCommands = []string{
	"ifconfig",
	"route -n",
	"arp -n",
	"cat /etc/hosts",
	"cat /etc/resolv.conf",
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
}

func confPageHandler(w http.ResponseWriter, r *http.Request) {
	cmd := r.FormValue("cmd")
	if cmd == "" {
		cmd = confCommands[0]
	} else {
		if !h.CheckValueInStrList(confCommands, cmd) {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
	}
	cmdfields := strings.Fields(cmd)

	if r.FormValue("data") == "1" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(h.ReadCommand(cmdfields[0], cmdfields[1:]...)))
	} else {
		ctx := &confPageContext{BasePageContext: app.NewBasePageContext("Network", "net", w, r)}
		app.AttachSubmenu(ctx.BasePageContext, "net", buildLocalMenu())
		ctx.SetMenuActive("conf")
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

func iptablesPageHandler(w http.ResponseWriter, r *http.Request) {
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
		ctx := &iptablesPageContext{BasePageContext: app.NewBasePageContext("Network", "net", w, r)}
		app.AttachSubmenu(ctx.BasePageContext, "net", buildLocalMenu())
		ctx.SetMenuActive("iptables")
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
		fields := strings.Fields(line)
		laddressDiv := strings.LastIndex(fields[3], ":")
		faddressDiv := strings.LastIndex(fields[4], ":")
		var state string
		var pidcmd []string
		pidcmdfield := fields[len(fields)-1]
		if pidcmdfield == "-" {
			pidcmd = []string{"", "-"}
		} else {
			pidcmd = strings.Split(pidcmdfield, "/")
		}
		if len(fields) == 7 {
			state = fields[5]
		} else if len(fields) != 6 {
			continue
		}
		result = append(result, []string{
			fields[0], fields[1], fields[2],
			fields[3][:laddressDiv], fields[3][laddressDiv+1:],
			fields[4][:faddressDiv], fields[4][faddressDiv+1:],
			state, pidcmd[0], pidcmd[1],
		})
	}
	return result, nil
}

var statusServCache = h.NewSimpleCache(1)

func statusServHandler(w http.ResponseWriter, r *http.Request) {
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

func actionHandler(w http.ResponseWriter, r *http.Request) {
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
