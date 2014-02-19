package logs

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"net/http"
	"strings"
)

// CreateRoutes for /users
func CreateRoutes(parentRoute *mux.Route) {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(mainPageHandler, "admin")).Name("users-index")
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	page := r.FormValue("sec")
	if page == "" {
		page = "who"
	}
	data := app.NewSimpleDataPageCtx(w, r, "Users", "", nil)
	data.Header1 = "Users"
	data.Tabs = []*app.MenuItem{
		app.NewMenuItemFromRoute("Who", "users-index").AddQuery("?sec=who").SetActve(page == "who"),
		app.NewMenuItemFromRoute("Users", "users-index").AddQuery("?sec=users").SetActve(page == "users"),
		app.NewMenuItemFromRoute("Groups", "users-index").AddQuery("?sec=groups").SetActve(page == "groups"),
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

	}
	app.RenderTemplateStd(w, data, "data.tmpl", "tabs.tmpl")
}
