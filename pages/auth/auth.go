package auth

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"k.prv/rpimon/app"
	"k.prv/rpimon/database"
	"k.prv/rpimon/helpers"
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

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	loginPageCtx := &LoginPageCtx{app.NewBasePageContext("Login", w, r),
		new(LoginForm), ""}
	app.RenderTemplate(w, loginPageCtx, "base", "login.tmpl", "flash.tmpl")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	loginPageCtx := &LoginPageCtx{app.NewBasePageContext("Login", w, r),
		new(LoginForm), ""}
	_ = r.ParseForm()
	values := r.Form
	err := decoder.Decode(loginPageCtx, values)
	if err != nil {
		l.Warn("Decode form error", err, values)
	}
	password := loginPageCtx.Password
	if password == "" || loginPageCtx.Login == "" {
		loginPageCtx.Message = "Missing login and/or password"
		app.RenderTemplate(w, loginPageCtx, "base", "login.tmpl", "flash.tmpl")
		return
	}
	user := database.GetUserByLogin(loginPageCtx.Login)
	if user != nil {
		cp_err := helpers.ComparePassword(user.Password, password)
		if cp_err != nil {
			loginPageCtx.Message = "Wrong user or password"
			app.RenderTemplate(w, loginPageCtx, "base", "login.tmpl", "flash.tmpl")
			return
		}
		l.Info("User %s log in", user.Login)
	}
	loginPageCtx.Set(app.USERID_SESSION, user.Login)
	loginPageCtx.SessionSave()
	if values["back"] != nil && values["back"][0] != "" {
		l.Debug("Redirect to ", values["back"][0])
		http.Redirect(w, r, values["back"][0], http.StatusFound)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func LogoffHandler(w http.ResponseWriter, r *http.Request) {
	session := app.GetSessionStore(w, r)
	session.Clear()
	session.Save(w, r)
	http.Redirect(w, r, "/", http.StatusFound)

}
