package modules

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/cfg"
	"k.prv/rpimon/app/context"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
)

var decoder = schema.NewDecoder()

// CreateRoutes for /main
func CreateRoutes(parentRoute *mux.Route) {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", context.HandleWithContextSec(mainPageHandler, "Main", "admin")).Name("m-pref-modules-index")
	subRouter.HandleFunc("/{module}", context.HandleWithContextSec(confModulePageHandler, "Main", "admin")).Name("m-pref-modules-module")
}

type (
	moduleSt struct {
		Name             string
		Title            string
		Description      string
		Enabled          bool
		ConfigurePageURL string
		Configurable     bool
	}

	modulesListForm struct {
		Modules []*moduleSt
	}

	pageCtx struct {
		*context.BasePageContext
		Form modulesListForm
	}
)

func mainPageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BasePageContext) {
	ctx := &pageCtx{BasePageContext: bctx}
	if r.Method == "POST" {
		r.ParseForm()
		if err := decoder.Decode(&ctx.Form, r.Form); err != nil {
			l.Warn("Decode form error", err, r.Form)
			return
		}
		for _, module := range ctx.Form.Modules {
			context.EnableModule(module.Name, module.Enabled)
		}
		context.SetMainMenu(ctx.BasePageContext)
		if err := cfg.SaveConfiguration(); err != nil {
			ctx.BasePageContext.AddFlashMessage("Saving configuration error: "+err.Error(),
				"error")
		} else {
			ctx.BasePageContext.AddFlashMessage("Configuration saved.", "success")
		}
		ctx.Save()
		http.Redirect(w, r, r.URL.String(), http.StatusFound)
		return
	}
	ctx.SetMenuActive("p-modules")
	for _, m := range context.GetModulesList() {
		if m.Internal {
			continue
		}
		ctx.Form.Modules = append(ctx.Form.Modules, &moduleSt{
			m.Name, m.Title, m.Description, m.Enabled(),
			m.ConfigurePageURL,
			m.Configurable,
		})
	}
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
		*context.BasePageContext
		Form   confModuleForm
		Module *context.Module
	}
)

func confModulePageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BasePageContext) {
	vars := mux.Vars(r)
	moduleName, _ := vars["module"]
	ctx := &confModulePageContext{BasePageContext: bctx}
	ctx.Module = context.GetModule(moduleName)
	if ctx.Module == nil {
		http.Error(w, "invalid module "+moduleName, http.StatusBadRequest)
		return
	}
	if !ctx.Module.Configurable {
		http.Error(w, "module not confiugrable", http.StatusBadRequest)
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
		if err := cfg.SaveConfiguration(); err != nil {
			ctx.BasePageContext.AddFlashMessage("Saving configuration error: "+err.Error(),
				"error")
		} else {
			ctx.BasePageContext.AddFlashMessage("Configuration saved.", "success")
		}
		ctx.Save()
		http.Redirect(w, r, app.GetNamedURL("module-index"), http.StatusFound)
		return
	}
	for key, val := range conf {
		if key != "enabled" {
			ctx.Form.Params = append(ctx.Form.Params, confParam{key, val})
		}
	}
	ctx.Form.Enabled = conf["enabled"] == "yes"
	ctx.SetMenuActive("p-modules")
	app.RenderTemplateStd(w, ctx, "pref/modules/conf.tmpl")
}
