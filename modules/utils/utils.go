package utils

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/logging"
	"net/http"
	"strconv"
	"strings"
)

var decoder = schema.NewDecoder()

// Module information
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
		Configurable: true,
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
	subRouter.HandleFunc("/", app.SecContext(mainPageHandler, "Utils", "admin")).Name("utils-index")
	subRouter.HandleFunc("/{group}/{command-id:[0-9]+}", app.SecContext(commandPageHandler, "Utils", "admin")).Name("utils-exec")
	subRouter.HandleFunc("/configure", app.SecContext(configurePageHandler, "Utils - Configuration", "admin")).Name("utils-conf")
	subRouter.HandleFunc("/configure/{group}", app.SecContext(confGroupPageHandler, "Utils - Configuration", "admin")).Name("utils-group")
	subRouter.HandleFunc("/configure/{group}/{util}", app.SecContext(confCommandPageHandler, "Utils - Configuration", "admin")).Name("utils-cmd")

	Module.ConfigurePageURL = app.GetNamedURL("utils-conf")

	return true
}
func getMenu(ctx *app.BaseCtx) (parentID string, menu *app.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		return "", nil
	}
	menu = app.NewMenuItemFromRoute("Utilities", "utils-index").SetID("utils").SetIcon("glyphicon glyphicon-wrench")
	return "", menu
}

type pageCtx struct {
	*app.DataPageCtx
	CurrentPage   string
	Configuration configuration
	Data          string
}

func mainPageHandler(r *http.Request, ctx *app.BaseCtx) {
	data := &pageCtx{
		DataPageCtx:   &app.DataPageCtx{BaseCtx: ctx},
		Configuration: config,
	}
	data.SetMenuActive("utils")
	ctx.RenderStd(data, "utils/utils.tmpl")
}

func commandPageHandler(r *http.Request, ctx *app.BaseCtx) {
	vars := mux.Vars(r)
	groupName, ok := vars["group"]
	if !ok || groupName == "" {
		l.Warn("page.utils commandPageHandler: missing group ", vars)
		mainPageHandler(r, ctx)
		return
	}

	group, ok := config.Utils[groupName]
	if !ok {
		l.Warn("page.utils commandPageHandler: wrong group ", vars)
		mainPageHandler(r, ctx)
		return
	}

	commandIDStr, ok := vars["command-id"]
	if !ok || commandIDStr == "" {
		l.Warn("page.utils commandPageHandler: wrong commandIDStr ", vars)
		mainPageHandler(r, ctx)
		return
	}

	commandID, err := strconv.Atoi(commandIDStr)
	if err != nil || commandID < 0 || commandID >= len(group) {
		l.Warn("page.utils commandPageHandler: wrong commandID ", vars)
		mainPageHandler(r, ctx)
		return
	}

	commandStr := group[commandID].Command
	command := strings.Split(commandStr, " ")

	result := h.ReadCommand(command[0], command[1:]...)
	if result == "" {
		result = "<b>Done</b> - No result"
	}
	w := ctx.ResponseWriter
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(result))
}

type (
	configurePageContext struct {
		*app.BaseCtx
		Utils map[string][]*utility
	}
)

func configurePageHandler(r *http.Request, bctx *app.BaseCtx) {
	ctx := &configurePageContext{BaseCtx: bctx}
	ctx.Utils = config.Utils
	ctx.RenderStd(ctx, "utils/conf-index.tmpl")
}

type (
	confGroupForm struct {
		Name string
	}

	confGroupPageContext struct {
		*app.BaseCtx
		Form confGroupForm
		New  bool
	}

	confCommandPageContext struct {
		*app.BaseCtx
		Form utility
		New  bool
	}
)

func (u *utility) validate() (errors []string) {
	u.Name = strings.TrimSpace(u.Name)
	u.Command = strings.TrimSpace(u.Command)
	if u.Name == "" {
		errors = append(errors, "Missing name")
	}
	if u.Command == "" {
		errors = append(errors, "Missing command")
	}
	return
}

func (f *confGroupForm) validate() (errors []string) {
	f.Name = strings.TrimSpace(f.Name)
	if f.Name == "" {
		errors = append(errors, "Missing name")
	}
	return
}

func confGroupPageHandler(r *http.Request, bctx *app.BaseCtx) {
	vars := mux.Vars(r)
	groupName, _ := vars["group"]
	ctx := &confGroupPageContext{BaseCtx: bctx}

	if r.Method == "POST" && r.FormValue("_method") != "" {
		r.Method = r.FormValue("_method")
	}
	switch r.Method {
	case "POST":
		r.ParseForm()
		if err := decoder.Decode(&ctx.Form, r.Form); err != nil {
			l.Warn("Decode form error", err, r.Form)
		}
		errors := ctx.Form.validate()
		if errors == nil || len(errors) == 0 {
			if ctx.Form.Name == groupName {
				// no changes
				ctx.Redirect(app.GetNamedURL("utils-conf"))
				return
			}
			if groupName == "<new>" {
				// new group
				config.Utils[ctx.Form.Name] = make([]*utility, 0)
			} else if _, found := config.Utils[ctx.Form.Name]; found {
				// group already exists - join
				config.Utils[ctx.Form.Name] = append(config.Utils[ctx.Form.Name], config.Utils[groupName]...)
				delete(config.Utils, groupName)
			} else {
				// rename group
				config.Utils[ctx.Form.Name] = config.Utils[groupName]
				delete(config.Utils, groupName)
			}
			if saveConf(ctx.BaseCtx) {
				ctx.Redirect(app.GetNamedURL("utils-conf"))
				return
			}
		} else {
			for _, err := range errors {
				ctx.BaseCtx.AddFlashMessage("Validation error: "+err, "error")
			}
			ctx.Save()
		}
	case "DELETE":
		delete(config.Utils, groupName)
		if saveConf(ctx.BaseCtx) {
			ctx.Redirect(app.GetNamedURL("utils-conf"))
			return
		}
	case "GET":
		ctx.New = groupName == "<new>"
		if !ctx.New {
			ctx.Form.Name = groupName
		}
	}
	ctx.Save()
	ctx.RenderStd(ctx, "utils/conf-group.tmpl")
}

func confCommandPageHandler(r *http.Request, bctx *app.BaseCtx) {
	vars := mux.Vars(r)
	groupName, _ := vars["group"]
	cmd, _ := vars["util"]
	ctx := &confCommandPageContext{BaseCtx: bctx}
	group := config.Utils[groupName]
	if r.Method == "POST" && r.FormValue("_method") != "" {
		r.Method = r.FormValue("_method")
	}
	switch r.Method {
	case "POST":
		r.ParseForm()
		if err := decoder.Decode(&ctx.Form, r.Form); err != nil {
			l.Warn("Decode form error", err, r.Form)
		}
		errors := ctx.Form.validate()
		if errors == nil || len(errors) == 0 {
			if group == nil {
				// create group when no exists
				group = make([]*utility, 0)
				config.Utils[groupName] = group
			}
			found := false
			for _, u := range group {
				if cmd == u.Name {
					u.Name = ctx.Form.Name
					u.Command = ctx.Form.Command
					found = true
					break
				}
			}
			if !found {
				// add new command
				group = append(group, &ctx.Form)
				config.Utils[groupName] = group
			}
			if saveConf(ctx.BaseCtx) {
				ctx.Redirect(app.GetNamedURL("utils-conf"))
				return
			}
		} else {
			for _, err := range errors {
				ctx.BaseCtx.AddFlashMessage("Validation error: "+err, "error")
			}
			ctx.Save()
		}
	case "DELETE":
		for idx, u := range group {
			if cmd == u.Name {
				config.Utils[groupName] = append(group[:idx], group[idx+1:]...)
				break
			}
		}
		if saveConf(ctx.BaseCtx) {
			ctx.Redirect(app.GetNamedURL("utils-conf"))
			return
		}

	case "GET":
		ctx.New = cmd == "<new>"
		if !ctx.New {
			for _, u := range config.Utils[groupName] {
				if cmd == u.Name {
					ctx.Form = *u
					break
				}
			}
		}
	}
	ctx.RenderStd(ctx, "utils/conf-cmd.tmpl")
}

// save configuration and add apriopriate flash message
func saveConf(bctx *app.BaseCtx) (success bool) {
	conf := Module.GetConfiguration()
	err := saveConfiguration(conf["config_file"])
	if err != nil {
		bctx.AddFlashMessage("Saving configuration error: "+err.Error(),
			"error")
	} else {
		bctx.AddFlashMessage("Configuration saved.", "success")
	}
	bctx.Save()
	return err == nil
}
