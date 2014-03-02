package users

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"k.prv/rpimon/app"
	"k.prv/rpimon/database"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
)

var decoder = schema.NewDecoder()

// CreateRoutes for /main
func CreateRoutes(parentRoute *mux.Route) {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.HandleWithContextSec(mainPageHandler, "Preferences - Users", "admin")).Name("p-users-index")
	subRouter.HandleFunc("/{user}", app.HandleWithContextSec(userPageHandler, "Preferences - User", "admin")).Name("p-users-user")
}

type (
	usersPageCtx struct {
		*app.BasePageContext
		Users []*database.User
	}
)

func mainPageHandler(w http.ResponseWriter, r *http.Request, bctx *app.BasePageContext) {
	ctx := &usersPageCtx{BasePageContext: bctx}
	ctx.Users = database.GetUsers()
	ctx.SetMenuActive("p-users")
	app.RenderTemplateStd(w, ctx, "pref/users/index.tmpl")
}

type (
	userForm struct {
		Login        string
		Privs        []string
		NewPassword  string
		NewPasswordC string
	}

	userPageCtx struct {
		*app.BasePageContext
		Form     userForm
		New      bool
		AllPrivs []string
	}
)

func (c *userPageCtx) HasPriv(perm string) bool {
	return app.CheckPermission(c.Form.Privs, perm)
}

func userPageHandler(w http.ResponseWriter, r *http.Request, bctx *app.BasePageContext) {
	vars := mux.Vars(r)
	login, _ := vars["user"]
	ctx := &userPageCtx{BasePageContext: bctx,
		AllPrivs: database.AllPrivs,
	}

	if login == "<new>" {
		ctx.New = true
	}

	if r.Method == "POST" && r.FormValue("_method") != "" {
		r.Method = r.FormValue("_method")
	}
	switch r.Method {
	case "POST":
		r.ParseForm()
		var err error
		if err = decoder.Decode(&ctx.Form, r.Form); err != nil {
			l.Warn("Decode form error", err, r.Form)
		}
		user := &database.User{
			Privs: ctx.Form.Privs,
		}
		if login == "<new>" {
			user.Login = ctx.Form.Login
			user.Password = database.CreatePassword(ctx.Form.NewPassword)
			if err = database.AddUser(user); err == nil {
				ctx.BasePageContext.AddFlashMessage("User added", "success")
				ctx.Save()
				http.Redirect(w, r, app.GetNamedURL("p-users-index"), http.StatusFound)
				return
			} else {
				ctx.AddFlashMessage("Add user errror: "+err.Error(), "error")
			}
		} else {
			// update user
			user.Login = login
			if ctx.Form.NewPassword != "" {
				user.Password = database.CreatePassword(ctx.Form.NewPassword)
			}
			if err = database.UpdateUser(user); err == nil {
				ctx.AddFlashMessage("User updated", "success")
				ctx.Save()
				http.Redirect(w, r, app.GetNamedURL("p-users-index"), http.StatusFound)
				return
			} else {
				ctx.AddFlashMessage("Update user errror: "+err.Error(), "error")
			}
		}
		if err != nil {
			ctx.AddFlashMessage("Error: "+err.Error(), "error")
		}

	case "DELETE":
		if err := database.DeleteUser(login); err == nil {
			ctx.AddFlashMessage("User deleted", "success")
			ctx.Save()
			http.Redirect(w, r, app.GetNamedURL("p-users-index"), http.StatusFound)
			return
		} else {
			ctx.AddFlashMessage("Update user errror: "+err.Error(), "error")
		}

	case "GET":
		if !ctx.New {
			user := database.GetUserByLogin(login)
			ctx.Form = userForm{
				Login: user.Login,
				Privs: user.Privs,
			}
		}
	}
	ctx.Save()
	app.RenderTemplateStd(w, ctx, "pref/users/user.tmpl")

}
