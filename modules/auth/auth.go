package auth

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"k.prv/rpimon/app"
	l "k.prv/rpimon/logging"
	"k.prv/rpimon/model"
	"net/http"
)

var decoder = schema.NewDecoder()

var subRouter *mux.Router

// Module information
var Module = &app.Module{
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
	subRouter.HandleFunc("/login", app.Context(loginPageHandler, "Login")).Name("auth-login")
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
		*app.BaseCtx
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

func loginPageHandler(r *http.Request, bctx *app.BaseCtx) {
	ctx := &loginPageCtx{bctx, new(loginForm), ""}
	if r.Method == "POST" {
		r.ParseForm()
		if err := decoder.Decode(ctx, r.Form); err != nil {
			l.Warn("Decode form error", err, r.Form)
			handleLoginError("Form error", ctx)
			return
		}
		if err := ctx.Validate(); err != "" {
			handleLoginError(err, ctx)
			return
		}
		user := model.GetUserByLogin(ctx.Login)
		if user == nil || !user.CheckPassword(ctx.Password) {
			handleLoginError("Wrong user or password", ctx)
			return
		}
		ctx.AddFlashMessage("User log in", "info")
		app.LoginUser(ctx.ResponseWriter, r, user)
		if back := r.FormValue("back"); back != "" {
			l.Debug("Redirect to ", back)
			ctx.Redirect(back)
		} else {
			ctx.Redirect("/")
		}
	} else {
		app.RenderTemplate(ctx.ResponseWriter, ctx, "login", "login.tmpl", "flash.tmpl")
	}
}

func handleLoginError(message string, ctx *loginPageCtx) {
	ctx.Message = message
	app.RenderTemplate(ctx.ResponseWriter, ctx, "login", "login.tmpl", "flash.tmpl")
}

func logoffHandler(w http.ResponseWriter, r *http.Request) {
	app.ClearSession(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
}
