package logs

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
	Name:          "system-users",
	Title:         "Users",
	Description:   "System users",
	AllPrivilages: nil,
	Init:          initModule,
	GetMenu:       getMenu,
}

// CreateRoutes for /users
func initModule(parentRoute *mux.Route) bool {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(mainPageHandler, "admin")).Name("users-index")
	return true
}

func getMenu(ctx *context.BaseCtx) (parentID string, menu *context.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		return "", nil
	}
	menu = app.NewMenuItemFromRoute("Users", "users-index").SetID("users").SetIcon("glyphicon glyphicon-user")
	return "system", menu
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	page := r.FormValue("sec")
	if page == "" {
		page = "who"
	}
	data := context.NewDataPageCtx(w, r, "Users")
	data.Header1 = "Users"
	data.Tabs = []*context.MenuItem{
		app.NewMenuItemFromRoute("Who", "users-index").AddQuery("?sec=who").SetActve(page == "who"),
		app.NewMenuItemFromRoute("Users", "users-index").AddQuery("?sec=users").SetActve(page == "users"),
		app.NewMenuItemFromRoute("Groups", "users-index").AddQuery("?sec=groups").SetActve(page == "groups"),
		app.NewMenuItemFromRoute("Last", "users-index").AddQuery("?sec=last").SetActve(page == "groups"),
	}
	data.SetMenuActive("users")
	switch page {
	case "who":
		data.Data = "WHO\n=========\n" + h.ReadCommand("who", "-a", "-H")
		data.Data += "\n\nW\n=========\n" + h.ReadCommand("w")
		data.Header2 = "Who"
	case "users":
		res, _ := h.ReadFile("/etc/passwd", -1)
		data.THead = []string{"Login", "x", "UID", "GUI", "Name", "Home", "Commnet"}
		for _, line := range strings.Split(res, "\n") {
			if line != "" {
				data.TData = append(data.TData, strings.Split(line, ":"))
			}
		}
		data.Header2 = "Users"
	case "groups":
		res, _ := h.ReadFile("/etc/group", -1)
		data.THead = []string{"Name", "x", "GUI", "Users"}
		for _, line := range strings.Split(res, "\n") {
			if line != "" {
				data.TData = append(data.TData, strings.Split(line, ":"))
			}
		}
		data.Header2 = "Groups"
	case "last":
		data.Data = "LAST\n=========\n" + h.ReadCommand("last")
		data.Data += "\n\nLASTB\n=========\n" + h.ReadCommand("sudo", "lastb")
		data.Header2 = "Last"
	}
	app.RenderTemplateStd(w, data, "data.tmpl", "tabs.tmpl")
}
