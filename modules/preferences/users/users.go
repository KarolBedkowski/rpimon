package users

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"k.prv/rpimon/app"
	l "k.prv/rpimon/logging"
	"k.prv/rpimon/model"
	"net/http"
)

var decoder = schema.NewDecoder()

// CreateRoutes for /main
func CreateRoutes(parentRoute *mux.Route) {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/users", app.SecContext(mainPageHandler, "Preferences - Users", "admin")).Name("m-pref-users-index")
	subRouter.HandleFunc("/users/{user}", app.SecContext(userPageHandler, "Preferences - User", "admin")).Name("m-pref-users-user")
	subRouter.HandleFunc("/profile", app.SecContext(profilePageHandler, "Profile", "")).Name("m-pref-user-profile")
}

type (
	usersPageCtx struct {
		*app.BaseCtx
		Users []*model.User
	}
)

func mainPageHandler(w http.ResponseWriter, r *http.Request, bctx *app.BaseCtx) {
	ctx := &usersPageCtx{BaseCtx: bctx}
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
		*app.BaseCtx
		Form     userForm
		New      bool
		AllPrivs map[string]app.Privilege
	}
)

func (c *userPageCtx) HasPriv(perm string) bool {
	return app.CheckPermission(c.Form.Privs, perm)
}

func userPageHandler(w http.ResponseWriter, r *http.Request, bctx *app.BaseCtx) {
	vars := mux.Vars(r)
	login, _ := vars["user"]
	ctx := &userPageCtx{BaseCtx: bctx,
		AllPrivs: app.AllPrivilages,
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
				ctx.BaseCtx.AddFlashMessage("User added", "success")
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
		*app.BaseCtx
		CPForm chgPassForm
		User   *model.User
	}
)

func profilePageHandler(w http.ResponseWriter, r *http.Request, bctx *app.BaseCtx) {
	ctx := &profileContext{
		BaseCtx: bctx,
		User:    model.GetUserByLogin(bctx.CurrentUser),
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
