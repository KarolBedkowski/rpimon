package modules

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"k.prv/rpimon/app"
	"k.prv/rpimon/cfg"
	l "k.prv/rpimon/logging"
	"net/http"
)

var decoder = schema.NewDecoder()

// CreateRoutes for /main
func CreateRoutes(parentRoute *mux.Route) {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.SecContext(mainPageHandler, "Main", "admin")).Name("m-pref-modules-index")
	subRouter.HandleFunc("/{module}", app.SecContext(confModulePageHandler, "Main", "admin")).Name("m-pref-modules-module")
}

type (
	moduleSt struct {
		Name             string
		Title            string
		Description      string
		Enabled          bool
		ConfigurePageURL string
		Configurable     bool
		Internal         bool
	}

	modulesListForm struct {
		Modules []*moduleSt
	}

	pageCtx struct {
		*app.BaseCtx
		Form modulesListForm
	}
)

func mainPageHandler(w http.ResponseWriter, r *http.Request, bctx *app.BaseCtx) {
	ctx := &pageCtx{BaseCtx: bctx}
	if r.Method == "POST" {
		r.ParseForm()
		if err := decoder.Decode(&ctx.Form, r.Form); err != nil {
			l.Warn("Decode form error", err, r.Form)
			return
		}
		for _, module := range ctx.Form.Modules {
			app.EnableModule(module.Name, module.Enabled)
		}
		app.SetMainMenu(ctx.BaseCtx)
		if err := cfg.SaveConfiguration(); err != nil {
			ctx.BaseCtx.AddFlashMessage("Saving configuration error: "+err.Error(),
				"error")
		} else {
			ctx.BaseCtx.AddFlashMessage("Configuration saved.", "success")
		}
		ctx.Save()
		http.Redirect(w, r, r.URL.String(), http.StatusFound)
		return
	}
	ctx.SetMenuActive("p-modules")
	for _, m := range app.GetModulesList() {
		ctx.Form.Modules = append(ctx.Form.Modules, &moduleSt{
			m.Name, m.Title, m.Description, m.Enabled(),
			m.ConfigurePageURL,
			m.Configurable,
			m.Internal,
		})
	}
	ctx.SetMenuActive("m-modules")
	app.RenderTemplateStd(w, ctx, "pref/modules/index.tmpl")
}

type (
	confParam struct {
		Name  string
		Value string
	}

	confModuleForm struct {
		Params  []confParam
		Enabled bool
	}

	confModulePageContext struct {
		*app.BaseCtx
		Form   confModuleForm
		Module *app.Module
	}
)

func confModulePageHandler(w http.ResponseWriter, r *http.Request, bctx *app.BaseCtx) {
	vars := mux.Vars(r)
	moduleName, _ := vars["module"]
	ctx := &confModulePageContext{BaseCtx: bctx}
	ctx.Module = app.GetModule(moduleName)
	if ctx.Module == nil {
		app.Render400(w, r, "Invalid module "+moduleName)
		return
	}
	if !ctx.Module.Configurable {
		app.Render400(w, r, "Module not configurable")
		return
	}
	conf := ctx.Module.GetConfiguration()
	ctx.SetMenuActive("modules")
	if r.Method == "POST" {
		r.ParseForm()
		if err := decoder.Decode(&ctx.Form, r.Form); err != nil {
			l.Warn("Decode form error", err, r.Form)
			return
		}
		for _, param := range ctx.Form.Params {
			conf[param.Name] = param.Value
		}
		if ctx.Form.Enabled {
			conf["enabled"] = "yes"
		} else {
			conf["enabled"] = "no"
		}
		ctx.Module.SaveConfiguration(conf)
		if err := cfg.SaveConfiguration(); err != nil {
			ctx.BaseCtx.AddFlashMessage("Saving configuration error: "+err.Error(),
				"error")
		} else {
			ctx.BaseCtx.AddFlashMessage("Configuration saved.", "success")
		}
		ctx.Save()
		http.Redirect(w, r, app.GetNamedURL("m-pref-modules-index"), http.StatusFound)
		return
	}
	for key, val := range conf {
		if key != "enabled" {
			ctx.Form.Params = append(ctx.Form.Params, confParam{key, val})
		}
	}
	ctx.Form.Enabled = conf["enabled"] == "yes"
	ctx.SetMenuActive("m-modules")
	app.RenderTemplateStd(w, ctx, "pref/modules/conf.tmpl")
}
