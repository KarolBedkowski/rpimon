package logs

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"net/http"
)

var subRouter *mux.Router

func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyLogged(mainPageHandler)).Name("users-index")
}

type PageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Data        string
}

func newNetPageCtx(w http.ResponseWriter, r *http.Request) *PageCtx {
	ctx := &PageCtx{BasePageContext: app.NewBasePageContext("Users", w, r)}
	ctx.CurrentMainMenuPos = "/users/"
	return ctx
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	data := newNetPageCtx(w, r)
	data.Data = h.ReadFromCommand("who", "-a")
	data.CurrentPage = "Who"
	app.RenderTemplate(w, data, "base", "base.tmpl", "log.tmpl", "flash.tmpl")
}
