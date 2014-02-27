package users

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"net/http"
	"strings"
)

var Module = &app.Module{
	Name:          "storage-smart",
	Title:         "Storage - SMART",
	Description:   "",
	AllPrivilages: nil,
	Init:          initModule,
	GetMenu:       getMenu,
}

// CreateRoutes for /storage
func initModule(parentRoute *mux.Route) bool {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(smartPageHandler, "admin")).Name("storage-smart")
	subRouter.HandleFunc("/serv/smart", app.VerifyPermission(servSmartHandler, "admin")).Name("storage-serv-smart")
	return true
}

func getMenu(ctx *app.BasePageContext) (parentId string, menu *app.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		return "", nil
	}
	menu = app.NewMenuItemFromRoute("SMART", "storage-smart").SetID("smart")
	return "storage", menu
}

type smartPageContext struct {
	*app.SimpleDataPageCtx
	Devices []string
}

func smartPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &smartPageContext{SimpleDataPageCtx: app.NewSimpleDataPageCtx(w, r, "Storage - SMART")}
	ctx.SetMenuActive("smart")
	for _, line := range strings.Split(h.ReadCommand("lsblk", "-r"), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, "disk") {
			fields := strings.Fields(line)
			ctx.Devices = append(ctx.Devices, fields[0])
		}
	}
	app.RenderTemplateStd(w, ctx, "storage/smart.tmpl")
}

func servSmartHandler(w http.ResponseWriter, r *http.Request) {
	dev := r.FormValue("dev")
	smart := h.ReadCommand("sudo", "smartctl", "--all", "/dev/"+dev)
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Write([]byte(smart))
}
