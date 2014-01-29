package logs

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"net/http"
)

// CreateRoutes for /users
func CreateRoutes(parentRoute *mux.Route) {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(mainPageHandler, "admin")).Name("users-index")
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.NewSimpleDataPageCtx(w, r, "Users", "users", "", nil)
	data.Data = "WHO\n=========\n" + h.ReadFromCommand("who", "-a", "-H")
	data.Data += "\n\nW\n=========\n" + h.ReadFromCommand("w")
	data.CurrentPage = "Who"
	app.RenderTemplate(w, data, "base", "base.tmpl", "data.tmpl", "flash.tmpl")
}
