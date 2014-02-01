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

// CreateRoutes for /pages
func CreateRoutes(parentRoute *mux.Route) {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(mainPageHandler, "admin")).Name("utils-index")
	subRouter.HandleFunc("/{group}/{command-id:[0-9]+}", app.VerifyPermission(commandPageHandler, "admin"))
}

type pageCtx struct {
	*app.SimpleDataPageCtx
	CurrentPage   string
	Configuration configuration
	Data          string
}

func newPageCtx(w http.ResponseWriter, r *http.Request) *pageCtx {
	ctx := &pageCtx{SimpleDataPageCtx: app.NewSimpleDataPageCtx(w, r, "Utils", "utils", "", nil)}
	ctx.Configuration = config
	return ctx
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	data := newPageCtx(w, r)
	app.RenderTemplateStd(w, data, "utils/utils.tmpl")
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

	commandIDStr, ok := vars["command-id"]
	if !ok || commandIDStr == "" {
		l.Warn("page.utils commandPageHandler: wrong commandIDStr ", vars)
		mainPageHandler(w, r)
		return
	}

	commandID, err := strconv.Atoi(commandIDStr)
	if err != nil || commandID < 0 || commandID >= len(group) {
		l.Warn("page.utils commandPageHandler: wrong commandID ", vars)
		mainPageHandler(w, r)
		return
	}

	commandStr := group[commandID].Command
	command := strings.Split(commandStr, " ")

	data := newPageCtx(w, r)
	data.CurrentPage = "Utils " + groupName + ": " + group[commandID].Name
	data.Data = h.ReadFromCommand(command[0], command[1:]...)
	app.RenderTemplateStd(w, data, "data.tmpl")
}
