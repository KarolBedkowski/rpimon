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

func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/login", loginPageHandler).Name("auth-login").Methods("GET")
	subRouter.HandleFunc("/login", loginHandler).Methods("POST")
	subRouter.HandleFunc("/logoff", LogoffHandler).Name("auth-logoff")
}

type LoginForm struct {
	Login    string
	Password string
	Message  string
}

type LoginPageCtx struct {
	*app.BasePageContext
	*LoginForm
	back string
}

func (ctx LoginPageCtx) Validate() (err string) {
	if ctx.Password == "" || ctx.Login == "" {
		return "Missing login and/or password"
	}
	return
}

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	loginPageCtx := &LoginPageCtx{app.NewBasePageContext("Login", w, r),
		new(LoginForm), ""}
	app.RenderTemplate(w, loginPageCtx, "base", "login.tmpl", "flash.tmpl")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	loginPageCtx := &LoginPageCtx{app.NewBasePageContext("Login", w, r),
		new(LoginForm), ""}
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
	l.Info("User %s log in", user.Login)
	loginPageCtx.Set(app.USERID_SESSION, user.Login)
	loginPageCtx.AddFlashMessage("User log in")
	loginPageCtx.SessionSave()
	if values["back"] != nil && values["back"][0] != "" {
		l.Debug("Redirect to ", values["back"][0])
		http.Redirect(w, r, values["back"][0], http.StatusFound)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func handleLoginError(message string, w http.ResponseWriter, ctx *LoginPageCtx) {
	ctx.Message = message
	app.RenderTemplate(w, ctx, "base", "login.tmpl", "flash.tmpl")
}

func LogoffHandler(w http.ResponseWriter, r *http.Request) {
	session := app.GetSessionStore(w, r)
	session.Clear()
	session.Save(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
}
