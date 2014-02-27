package utils

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"strconv"
	"strings"
)

var Module *app.Module

func init() {
	Module = &app.Module{
		Name:          "utilities",
		Title:         "Utilities",
		Description:   "Various utilities",
		AllPrivilages: nil,
		Init:          initModule,
		GetMenu:       getMenu,
		Defaults: map[string]string{
			"config_file": "./utils.json",
		},
	}
}

// CreateRoutes for /pages
func initModule(parentRoute *mux.Route) bool {
	conf := Module.GetConfiguration()
	if err := loadConfiguration(conf["config_file"]); err != nil {
		l.Warn("Utils: failed load configuration file: %s", err)
		return false
	}
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.HandleWithContextSec(mainPageHandler, "Utils", "admin")).Name("utils-index")
	subRouter.HandleFunc("/{group}/{command-id:[0-9]+}", app.HandleWithContextSec(commandPageHandler, "Utils", "admin"))
	return true
}
func getMenu(ctx *app.BasePageContext) (parentId string, menu *app.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		return "", nil
	}
	menu = app.NewMenuItemFromRoute("Utilities", "utils-index").SetID("utils").SetIcon("glyphicon glyphicon-wrench")
	return "", menu
}

type pageCtx struct {
	*app.SimpleDataPageCtx
	CurrentPage   string
	Configuration configuration
	Data          string
}

func mainPageHandler(w http.ResponseWriter, r *http.Request, ctx *app.BasePageContext) {
	data := &pageCtx{
		SimpleDataPageCtx: &app.SimpleDataPageCtx{BasePageContext: ctx},
		Configuration:     config,
	}
	data.SetMenuActive("utils")
	app.RenderTemplateStd(w, data, "utils/utils.tmpl")
}

func commandPageHandler(w http.ResponseWriter, r *http.Request, ctx *app.BasePageContext) {
	vars := mux.Vars(r)
	groupName, ok := vars["group"]
	if !ok || groupName == "" {
		l.Warn("page.utils commandPageHandler: missing group ", vars)
		mainPageHandler(w, r, ctx)
		return
	}

	group, ok := config.Utils[groupName]
	if !ok {
		l.Warn("page.utils commandPageHandler: wrong group ", vars)
		mainPageHandler(w, r, ctx)
		return
	}

	commandIDStr, ok := vars["command-id"]
	if !ok || commandIDStr == "" {
		l.Warn("page.utils commandPageHandler: wrong commandIDStr ", vars)
		mainPageHandler(w, r, ctx)
		return
	}

	commandID, err := strconv.Atoi(commandIDStr)
	if err != nil || commandID < 0 || commandID >= len(group) {
		l.Warn("page.utils commandPageHandler: wrong commandID ", vars)
		mainPageHandler(w, r, ctx)
		return
	}

	commandStr := group[commandID].Command
	command := strings.Split(commandStr, " ")

	data := &pageCtx{
		SimpleDataPageCtx: &app.SimpleDataPageCtx{BasePageContext: ctx},
		Configuration:     config,
	}
	data.CurrentPage = "Utils " + groupName + ": " + group[commandID].Name
	data.Data = h.ReadCommand(command[0], command[1:]...)
	app.RenderTemplateStd(w, data, "data.tmpl")
}
