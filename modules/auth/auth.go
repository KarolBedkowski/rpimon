package auth

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	"k.prv/rpimon/app/session"
	l "k.prv/rpimon/logging"
	"k.prv/rpimon/model"
	"net/http"
)

var decoder = schema.NewDecoder()

var subRouter *mux.Router

// Module information
var Module = &context.Module{
	Name:          "auth",
	Title:         "Authentication",
	Description:   "",
	AllPrivilages: nil,
	Init:          initModule,
	Internal:      true,
}

// CreateRoutes for /auth
func initModule(parentRoute *mux.Route) bool {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/login", context.HandleWithContext(loginPageHandler, "Login")).Name("auth-login")
	subRouter.HandleFunc("/logoff", logoffHandler).Name("auth-logoff")
	return true
}

type (
	loginForm struct {
		Login    string
		Password string
		Message  string
	}

	loginPageCtx struct {
		*context.BasePageContext
		*loginForm
		back string
	}
)

func (ctx loginPageCtx) Validate() (err string) {
	if ctx.Password == "" || ctx.Login == "" {
		return "Missing login and/or password"
	}
	return
}

func loginPageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BasePageContext) {
	ctx := &loginPageCtx{bctx, new(loginForm), ""}
	if r.Method == "POST" {
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
		user := model.GetUserByLogin(ctx.Login)
		if user == nil || !user.CheckPassword(ctx.Password) {
			handleLoginError("Wrong user or password", w, ctx)
			return
		}
		ctx.AddFlashMessage("User log in", "info")
		app.LoginUser(w, r, user)
		if back := r.FormValue("back"); back != "" {
			l.Debug("Redirect to ", back)
			http.Redirect(w, r, back, http.StatusFound)
		} else {
			http.Redirect(w, r, "/", http.StatusFound)
		}
	} else {
		app.RenderTemplate(w, ctx, "login", "login.tmpl", "flash.tmpl")
	}
}

func handleLoginError(message string, w http.ResponseWriter, ctx *loginPageCtx) {
	ctx.Message = message
	app.RenderTemplate(w, ctx, "login", "login.tmpl", "flash.tmpl")
}

func logoffHandler(w http.ResponseWriter, r *http.Request) {
	session.ClearSession(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
}
