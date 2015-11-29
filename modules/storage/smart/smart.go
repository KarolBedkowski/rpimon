package users

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	h "k.prv/rpimon/helpers"
	"net/http"
	"strings"
)

// Module information
var Module = &context.Module{
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

func getMenu(ctx *context.BasePageContext) (parentID string, menu *context.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		return "", nil
	}
	menu = app.NewMenuItemFromRoute("SMART", "storage-smart").SetID("smart")
	return "storage", menu
}

type smartPageContext struct {
	*context.DataPageCtx
	Devices []string
}

func smartPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &smartPageContext{DataPageCtx: context.NewDataPageCtx(w, r, "Storage - SMART")}
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
