package users

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	l "k.prv/rpimon/helpers/logging"
	"k.prv/rpimon/model"
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
		Users []*model.User
	}
)

func mainPageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BasePageContext) {
	ctx := &usersPageCtx{BasePageContext: bctx}
	ctx.Users = model.GetAllUsers()
	ctx.SetMenuActive("m-users")
	app.RenderTemplateStd(w, ctx, "pref/users/index.tmpl")
}

type (
	userForm struct {
		*model.User
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
		ctx.Form.User = &model.User{}
	} else {
		ctx.Form.User = model.GetUserByLogin(login)
	}

	if r.Method == "POST" && r.FormValue("_method") != "" {
		r.Method = r.FormValue("_method")
	}
	var err error
	switch r.Method {
	case "POST":
		r.ParseForm()
		if err = decoder.Decode(&ctx.Form, r.Form); err != nil {
			l.Warn("Decode form error", err, r.Form)
		}
		if login == "<new>" {
			ctx.Form.User.UpdatePassword(ctx.Form.NewPassword)
			if err = model.AddUser(ctx.Form.User); err == nil {
				ctx.BasePageContext.AddFlashMessage("User added", "success")
				ctx.Save()
				http.Redirect(w, r, app.GetNamedURL("m-pref-users-index"), http.StatusFound)
				return
			}
			ctx.AddFlashMessage("Add user errror: "+err.Error(), "error")
		} else {
			// update user
			if ctx.Form.NewPassword != "" {
				ctx.Form.User.UpdatePassword(ctx.Form.NewPassword)
			}
			if err = model.UpdateUser(ctx.Form.User); err == nil {
				ctx.AddFlashMessage("User updated", "success")
				ctx.Save()
				http.Redirect(w, r, app.GetNamedURL("m-pref-users-index"), http.StatusFound)
				return
			}
			ctx.AddFlashMessage("Update user errror: "+err.Error(), "error")
		}
		if err != nil {
			ctx.AddFlashMessage("Error: "+err.Error(), "error")
		}

	case "DELETE":
		if err = model.DeleteUser(login); err == nil {
			ctx.AddFlashMessage("User deleted", "success")
			ctx.Save()
			http.Redirect(w, r, app.GetNamedURL("m-pref-users-index"), http.StatusFound)
			return
		}
		ctx.AddFlashMessage("Update user errror: "+err.Error(), "error")
	}
	ctx.SetMenuActive("m-users")
	ctx.Save()
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
		User   *model.User
	}
)

func profilePageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BasePageContext) {
	ctx := &profileContext{
		BasePageContext: bctx,
		User:            model.GetUserByLogin(bctx.CurrentUser),
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
				if err = model.UpdateUser(ctx.User); err == nil {
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
	ctx.SetMenuActive("m-user-profile")
	app.RenderTemplateStd(w, ctx, "pref/users/profile.tmpl")
}
