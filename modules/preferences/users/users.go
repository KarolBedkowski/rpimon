package users

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/cfg"
	"k.prv/rpimon/app/context"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
)

var decoder = schema.NewDecoder()

// CreateRoutes for /main
func CreateRoutes(parentRoute *mux.Route) {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/users", context.HandleWithContextSec(mainPageHandler, "Preferences - Users", "admin")).Name("m-pref-users-index")
	subRouter.HandleFunc("/users/{user}", context.HandleWithContextSec(userPageHandler, "Preferences - User", "admin")).Name("m-pref-users-user")
	subRouter.HandleFunc("/profile", context.HandleWithContextSec(profilePageHandler, "Profile", "")).Name("m-pref-user-profile")
}

type (
	usersPageCtx struct {
		*context.BasePageContext
		Users []*cfg.User
	}
)

func mainPageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BasePageContext) {
	ctx := &usersPageCtx{BasePageContext: bctx}
	ctx.Users = cfg.GetAllUsers()
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
		*context.BasePageContext
		Form     userForm
		New      bool
		AllPrivs map[string]context.Privilege
	}
)

func (c *userPageCtx) HasPriv(perm string) bool {
	return app.CheckPermission(c.Form.Privs, perm)
}

func userPageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BasePageContext) {
	vars := mux.Vars(r)
	login, _ := vars["user"]
	ctx := &userPageCtx{BasePageContext: bctx,
		AllPrivs: context.AllPrivilages,
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
		user := &cfg.User{
			Privs: ctx.Form.Privs,
		}
		if login == "<new>" {
			user.Login = ctx.Form.Login
			user.UpdatePassword(ctx.Form.NewPassword)
			if err = cfg.AddUser(user); err == nil {
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
				user.UpdatePassword(ctx.Form.NewPassword)
			}
			if err = cfg.UpdateUser(user); err == nil {
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
		if err := cfg.DeleteUser(login); err == nil {
			ctx.AddFlashMessage("User deleted", "success")
			ctx.Save()
			http.Redirect(w, r, app.GetNamedURL("p-users-index"), http.StatusFound)
			return
		} else {
			ctx.AddFlashMessage("Update user errror: "+err.Error(), "error")
		}

	case "GET":
		if !ctx.New {
			user := cfg.GetUserByLogin(login)
			ctx.Form = userForm{
				Login: user.Login,
				Privs: user.Privs,
			}
		}
	}
	ctx.Save()
	ctx.SetMenuActive("p-users")
	app.RenderTemplateStd(w, ctx, "pref/users/user.tmpl")

}

type (
	chgPassForm struct {
		ChangePass   bool
		OldPassword  string
		NewPassword  string
		NewPasswordC string
	}

	profileContext struct {
		*context.BasePageContext
		CPForm chgPassForm
		User   *cfg.User
	}
)

func profilePageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BasePageContext) {
	ctx := &profileContext{
		BasePageContext: bctx,
		User:            cfg.GetUserByLogin(bctx.CurrentUser),
	}
	switch r.Method {
	case "POST":
		r.ParseForm()
		var err error
		if err = decoder.Decode(&ctx.CPForm, r.Form); err != nil {
			l.Warn("Decode form error", err, r.Form)
		}
		if ctx.CPForm.ChangePass {
			if ctx.User.CheckPassword(ctx.CPForm.OldPassword) {
				ctx.User.UpdatePassword(ctx.CPForm.NewPassword)
				if err = cfg.UpdateUser(ctx.User); err == nil {
					ctx.AddFlashMessage("User updated", "success")
				} else {
					ctx.AddFlashMessage("Update user errror: "+err.Error(), "error")
				}
			} else {
				l.Info("Change password for user error: wrong password")
				ctx.AddFlashMessage("Wrong old password", "error")
			}
		}
	}
	ctx.Save()
	ctx.SetMenuActive("p-user-profile")
	app.RenderTemplateStd(w, ctx, "pref/users/profile.tmpl")
}
