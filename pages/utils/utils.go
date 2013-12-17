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

var subRouter *mux.Router

func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyLogged(mainPageHandler)).Name("utils-index")
	subRouter.HandleFunc("/{group}/{command-id:[0-9]+}", app.VerifyLogged(commandPageHandler))
}

type pageCtx struct {
	*app.BasePageContext
	CurrentPage   string
	Configuration configuration
	Data          string
}

func newPageCtx(w http.ResponseWriter, r *http.Request) *pageCtx {
	ctx := &pageCtx{BasePageContext: app.NewBasePageContext("Utils", w, r)}
	ctx.CurrentMainMenuPos = "/utils/"
	ctx.Configuration = config
	return ctx
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	data := newPageCtx(w, r)
	app.RenderTemplate(w, data, "base", "base.tmpl", "utils/utils.tmpl", "flash.tmpl")
}

func commandPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupName, ok := vars["group"]
	if !ok || groupName == "" {
		l.Warn("page.utils commandPageHandler: missing group ", vars)
		mainPageHandler(w, r)
		return
	}

	group, ok := config.Utils[groupName]
	if !ok {
		l.Warn("page.utils commandPageHandler: wrong group ", vars)
		mainPageHandler(w, r)
		return
	}

	commandIdStr, ok := vars["command-id"]
	if !ok || commandIdStr == "" {
		l.Warn("page.utils commandPageHandler: wrong commandIdStr ", vars)
		mainPageHandler(w, r)
		return
	}

	commandId, err := strconv.Atoi(commandIdStr)
	if err != nil || commandId < 0 || commandId >= len(group) {
		l.Warn("page.utils commandPageHandler: wrong commandId ", vars)
		mainPageHandler(w, r)
		return
	}

	commandStr := group[commandId].Command
	command := strings.Split(commandStr, " ")

	data := newPageCtx(w, r)
	data.CurrentPage = "Utils " + groupName + ": " + group[commandId].Name
	data.Data = h.ReadFromCommand(command[0], command[1:]...)
	app.RenderTemplate(w, data, "base", "base.tmpl", "log.tmpl", "flash.tmpl")
}
