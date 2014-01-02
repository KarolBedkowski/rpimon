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
	loginPageCtx := &loginPageCtx{app.NewBasePageContext("Login", "auth-login", w, r),
		new(loginForm), ""}
	app.RenderTemplate(w, loginPageCtx, "base", "login.tmpl", "flash.tmpl")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	loginPageCtx := &loginPageCtx{app.NewBasePageContext("Login", "auth-login", w, r),
		new(loginForm), ""}
	r.ParseForm()
	values := r.Form
	if err := decoder.Decode(loginPageCtx, values); err != nil {
		l.Warn("Decode form error", err, values)
		handleLoginError("Form error", w, loginPageCtx)
		return
	}
	if err := loginPageCtx.Validate(); err != "" {
		handleLoginError(err, w, loginPageCtx)
		return
	}
	user := database.GetUserByLogin(loginPageCtx.Login)
	if user == nil || !app.ComparePassword(user.Password, loginPageCtx.Password) {
		handleLoginError("Wrong user or password", w, loginPageCtx)
		return
	}
	loginPageCtx.AddFlashMessage("User log in")
	app.LoginUser(w, r, user.Login)
	if values["back"] != nil && values["back"][0] != "" {
		l.Debug("Redirect to ", values["back"][0])
		http.Redirect(w, r, values["back"][0], http.StatusFound)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func handleLoginError(message string, w http.ResponseWriter, ctx *loginPageCtx) {
	ctx.Message = message
	app.RenderTemplate(w, ctx, "base", "login.tmpl", "flash.tmpl")
}

func logoffHandler(w http.ResponseWriter, r *http.Request) {
	app.ClearSession(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
}
