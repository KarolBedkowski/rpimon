package modules

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"k.prv/rpimon/app"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
)

var decoder = schema.NewDecoder()

// CreateRoutes for /main
func CreateRoutes(parentRoute *mux.Route) {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.HandleWithContextSec(mainPageHandler, "Main", "admin")).Name("modules-index")
}

type (
	modulesListForm struct {
		Modules []*app.Module
	}

	pageCtx struct {
		*app.BasePageContext
		Form modulesListForm
	}
)

func mainPageHandler(w http.ResponseWriter, r *http.Request, bctx *app.BasePageContext) {
	ctx := &pageCtx{BasePageContext: bctx}
	ctx.SetMenuActive("modules")

	if r.Method == "POST" {
		r.ParseForm()
		if err := decoder.Decode(&ctx.Form, r.Form); err != nil {
			l.Warn("Decode form error", err, r.Form)
		}
		for _, module := range ctx.Form.Modules {
			app.SetModuleEnabled(module.Name, module.Enabled)
		}
		app.SetMainMenu(ctx.BasePageContext)
	}
	ctx.Form.Modules = app.GetModulesList()
	app.RenderTemplateStd(w, ctx, "pref/modules/index.tmpl")
}
