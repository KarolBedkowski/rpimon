package users

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"strings"
)

var subRouter *mux.Router

// CreateRoutes for /net
func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(mainPageHandler, "admin")).Name("net-index")
	subRouter.HandleFunc("/{page}", app.VerifyPermission(mainPageHandler, "admin")).Name("net-page")
}

var localMenu []*app.MenuItem

func createLocalMenu() []*app.MenuItem {
	if localMenu == nil {

		localMenu = []*app.MenuItem{app.NewMenuItemFromRoute("IFConfig", "net-page", "page", "ifconfig").SetID("ifconfig"),
			app.NewMenuItemFromRoute("IPTables", "net-page", "page", "iptables").SetID("iptables"),
			app.NewMenuItemFromRoute("Netstat", "net-page", "page", "netstat").SetID("netstat"),
			app.NewMenuItemFromRoute("Conenctions", "net-page", "page", "connenctions").SetID("connenctions")}
	}
	return localMenu
}

type networkPageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Data        string
	Connections [][]string
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	data := &networkPageCtx{BasePageContext: app.NewBasePageContext("Network", "net", w, r)}
	data.LocalMenu = createLocalMenu()
	vars := mux.Vars(r)
	page, ok := vars["page"]
	if !ok {
		page = "ifconfig"
	}
	data.CurrentLocalMenuPos = page
	data.CurrentPage = page
	switch page {
	case "ifconfig":
		data.Data = h.ReadFromCommand("ip", "addr")
	case "iptables":
		data.Data = h.ReadFromCommand("sudo", "iptables", "-L", "-vn")
	case "netstat":
		data.Connections, _ = getConnections("sudo", "netstat", "-lpn", "--inet", "--inet6")
		app.RenderTemplate(w, data, "base", "base.tmpl", "net/netstat.tmpl", "flash.tmpl")
		return
	case "connenctions":
		data.Connections, _ = getConnections("sudo", "netstat", "-pn", "--inet", "--inet6")
		app.RenderTemplate(w, data, "base", "base.tmpl", "net/netstat.tmpl", "flash.tmpl")
		return
	}
	app.RenderTemplate(w, data, "base", "base.tmpl", "log.tmpl", "flash.tmpl")
}

func getConnections(command string, args ...string) ([][]string, error) {
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
		l.Debug("%#v, %v %v", fields, laddressDiv, faddressDiv)
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
