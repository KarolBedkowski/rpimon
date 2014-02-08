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
	subRouter.HandleFunc("/serv/info", app.VerifyPermission(infoHandler, "admin")).Name("net-serv-info")
	subRouter.HandleFunc("/{page}", app.VerifyPermission(subPageHandler, "admin")).Name("net-page")
	localMenu = []*app.MenuItem{
		app.NewMenuItemFromRoute("Status", "net-index").SetID("status"),
		app.NewMenuItemFromRoute("Configuration", "net-conf").SetID("conf"),
		app.NewMenuItemFromRoute("IPTables", "net-page", "page", "iptables").SetID("iptables"),
		app.NewMenuItemFromRoute("Netstat", "net-page", "page", "netstat").SetID("netstat"),
		app.NewMenuItemFromRoute("Conenctions", "net-page", "page", "connenctions").SetID("connenctions")}
}

var localMenu []*app.MenuItem

type mainPageContext struct {
	*app.BasePageContext
	Interfaces *monitor.InterfacesStruct
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &mainPageContext{BasePageContext: app.NewBasePageContext("Network", "net", w, r)}
	ctx.CurrentLocalMenuPos = "status"
	ctx.LocalMenu = localMenu
	ctx.Interfaces = monitor.GetInterfacesInfo()
	app.RenderTemplateStd(w, ctx, "net/status.tmpl")
}

func subPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page, ok := vars["page"]
	if !ok {
		http.Redirect(w, r, app.GetNamedURL("net-index"), http.StatusFound)
	}
	data := app.NewSimpleDataPageCtx(w, r, "Network", "net", page, localMenu)
	data.CurrentLocalMenuPos = page
	switch page {
	case "iptables":
		data.Data = h.ReadFromCommand("sudo", "iptables", "-L", "-vn")
	case "netstat":
		data.THead = []string{"Proto", "Recv-Q", "Send-Q", "Local Address", "Port", "Foreign Address", "Port", "State", "PID", "Program name"}
		data.TData, _ = netstat("sudo", "netstat", "-lpn", "--inet", "--inet6")
	case "connenctions":
		data.THead = []string{"Proto", "Recv-Q", "Send-Q", "Local Address", "Port", "Foreign Address", "Port", "State", "PID", "Program name"}
		data.TData, _ = netstat("sudo", "netstat", "-pn", "--inet", "--inet6")
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
		ok := false
		for _, dcmd := range confCommands {
			if cmd == dcmd {
				ok = true
				break
			}
		}
		if !ok {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
	}
	cmdfields := strings.Fields(cmd)

	if r.FormValue("data") == "1" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(h.ReadFromCommand(cmdfields[0], cmdfields[1:]...)))
	} else {
		ctx := &confPageContext{BasePageContext: app.NewBasePageContext("Network", "net", w, r)}
		ctx.CurrentLocalMenuPos = "conf"
		ctx.LocalMenu = localMenu
		ctx.Current = cmd
		ctx.Commands = &confCommands
		ctx.Data = h.ReadFromCommand(cmdfields[0], cmdfields[1:]...)
		app.RenderTemplateStd(w, ctx, "net/conf.tmpl")
	}
}

func netstat(command string, args ...string) ([][]string, error) {
	result := make([][]string, 0)
	res := h.ReadFromCommand(command, args...)
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

var infoHandlerCache = h.NewSimpleCache(1)

func infoHandler(w http.ResponseWriter, r *http.Request) {
	data := infoHandlerCache.Get(func() h.Value {
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
