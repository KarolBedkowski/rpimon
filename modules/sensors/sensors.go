package sensors

import (
	"bufio"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"net/http"
	"os"
	"strings"
)

// Module information
var Module = &app.Module{
	Name:          "sensors",
	Title:         "Sensors",
	Description:   "",
	AllPrivilages: nil,
	Init:          initModule,
	GetMenu:       getMenu,
}

// CreateRoutes for /users
func initModule(parentRoute *mux.Route) bool {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.SecContext(mainPageHandler, "Sensors", "")).Name("sensors-index")
	return true
}

func getMenu(ctx *app.BaseCtx) (parentID string, menu *app.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "") {
		return "", nil
	}
	menu = app.NewMenuItemFromRoute("Sensors", "sensors-index").SetID("sensors.index").SetIcon("glyphicon glyphicon-cog")
	return "", menu
}

func mainPageHandler(r *http.Request, bctx *app.BaseCtx) {
	page := r.FormValue("sec")
	if page == "" {
		page = "w1"
	}
	data := &app.DataPageCtx{BaseCtx: bctx}
	data.Header1 = "Sensors"
	data.Tabs = []*app.MenuItem{
		app.NewMenuItemFromRoute("1-Wire", "other-index").AddQuery("?sec=w1").SetActve(page == "w1"),
	}
	switch page {
	case "w1":
		data.Data = getW1Data()
		data.Header2 = "1-Wire"
	}
	data.SetMenuActive("sensors-index")
	bctx.RenderStd(data, "data.tmpl", "tabs.tmpl")
}

func getW1Data() string {
	var sensors []string
	file, err := os.Open("/sys/bus/w1/devices/w1_bus_master1/w1_master_slaves")
	if err != nil {
		return "Error: " + err.Error()
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		if line, err := reader.ReadString('\n'); err == nil {
			line = strings.Trim(line, " \n")
			if len(line) > 0 {
				sensors = append(sensors, line)
			}
		} else {
			break
		}
	}

	if len(sensors) == 0 {
		return "Not found any 1-Wire slaves"
	}

	var outp []string

	for _, id := range sensors {
		outp = append(outp, id, "================================")
		if val, err := h.ReadFile("/sys/bus/w1/devices/"+id+"/w1_slave", -1); err == nil {
			outp = append(outp, val, "------------", "")
		} else {
			outp = append(outp, err.Error(), "------------", "")
		}
	}

	return strings.Join(outp, "\n")
}
