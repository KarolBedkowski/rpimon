package auth

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"k.prv/rpimon/app"
	"k.prv/rpimon/database"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
)

var decoder = schema.NewDecoder()

var subRouter *mux.Router

// CreateRoutes for /auth
func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/login", loginPageHandler).Name("auth-login").Methods("GET")
	subRouter.HandleFunc("/login", loginHandler).Methods("POST")
	subRouter.HandleFunc("/logoff", logoffHandler).Name("auth-logoff")
}

type loginForm struct {
	Login    string
	Password string
	Message  string
}

type loginPageCtx struct {
	*app.BasePageContext
	*loginForm
	back string
}

func (ctx loginPageCtx) Validate() (err string) {
	if ctx.Password == "" || ctx.Login == "" {
		return "Missing login and/or password"
	}
	return
}

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &loginPageCtx{app.NewBasePageContext("Login", "auth-login", w, r),
		new(loginForm), ""}
	app.RenderTemplate(w, ctx, "login", "login.tmpl", "flash.tmpl")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &loginPageCtx{app.NewBasePageContext("Login", "auth-login", w, r),
		new(loginForm), ""}
	r.ParseForm()
	if err := decoder.Decode(ctx, r.Form); err != nil {
		l.Warn("Decode form error", err, r.Form)
		handleLoginError("Form error", w, ctx)
		return
	}
	if err := ctx.Validate(); err != "" {
		handleLoginError(err, w, ctx)
		return
	}
	user := database.GetUserByLogin(ctx.Login)
	if user == nil || !app.ComparePassword(user.Password, ctx.Password) {
		handleLoginError("Wrong user or password", w, ctx)
		return
	}
	ctx.AddFlashMessage("User log in", "info")
	app.LoginUser(w, r, user.Login)
	if back := r.FormValue("back"); back != "" {
		l.Debug("Redirect to ", back)
		http.Redirect(w, r, back, http.StatusFound)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func handleLoginError(message string, w http.ResponseWriter, ctx *loginPageCtx) {
	ctx.Message = message
	app.RenderTemplate(w, ctx, "login", "login.tmpl", "flash.tmpl")
}

func logoffHandler(w http.ResponseWriter, r *http.Request) {
	app.ClearSession(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
}
