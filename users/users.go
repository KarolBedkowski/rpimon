package users

import (
	"../app"
	"../database"
	"../helpers"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

var subRouter *mux.Router

func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", usersHandler).Name("users-list")
	subRouter.HandleFunc("/{id:[0-9]+}", editUserHandler)
}

type UsersPageCtx struct {
	*app.BasePageContext
	Users []database.User
}

func newUsersPageCtx(w http.ResponseWriter, r *http.Request, users []database.User) *UsersPageCtx {
	return &UsersPageCtx{app.NewBasePageContext("Users", w, r), users}
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	if GetLoggedUser(w, r, true) == nil {
		return
	}
	a := context.Get(r, "APP").(*app.WebApp)
	data := newUsersPageCtx(w, r, database.UsersList())
	a.RenderTemplate(w, "base", data, "base.tmpl", "users/users.tmpl")
}

type UserForm struct {
	database.User
	Password1 string
	Password2 string
}

type EditPageCtx struct {
	*app.BasePageContext
	*UserForm
	Message   string
	CsrfToken string
}

func newEditPageCtx(w http.ResponseWriter, r *http.Request, msg string) *EditPageCtx {
	return &EditPageCtx{app.NewBasePageContext("User", w, r),
		&UserForm{}, msg, ""}
}

func (form *UserForm) validate() (err string) {
	err = ""
	if form.Login == "" {
		err = "Missing login"
		return
	}
	if form.Name == "" {
		err = "Missing name"
		return
	}
	if form.Password1 != "" {
		if form.Password1 != form.Password2 {
			err = "Mismatch passwords"
		}
	}
	return
}

func editUserHandler(w http.ResponseWriter, r *http.Request) {
	if GetLoggedUser(w, r, true) == nil {
		return
	}
	a := context.Get(r, "APP").(*app.WebApp)
	editPage := newEditPageCtx(w, r, "")
	editPage.CsrfToken = context.Get(r, app.CONTEXT_CSRF_TOKEN).(string)
	switch r.Method {
	case "GET":
		{
			vars := mux.Vars(r)
			userId, ok := vars["id"]
			if ok && userId != "" {
				userIdI, _ := strconv.ParseInt(userId, 10, 64)
				editPage.User = *database.GetUserById(userIdI)
			}
			a.RenderTemplate(w, "base", editPage, "base.tmpl", "users/edit.tmpl")
			return
		}
	case "POST":
		{
			r.ParseForm()
			err := decoder.Decode(editPage, r.Form)
			if err != nil {
				log.Print("Decoding form error", err)
			}
			msg := editPage.validate()
			log.Print("Validate: ", msg)
			if msg != "" {
				editPage.Message = msg
				a.RenderTemplate(w, "base", editPage, "base.tmpl", "users/edit.tmpl")
				return
			}
			var user *database.User
			if editPage.Id > 0 {
				user = database.GetUserById(editPage.Id)
			} else {
				user = new(database.User)
			}
			// Copy
			user.Login = editPage.Login
			user.Name = editPage.Name
			if editPage.Password1 != "" {
				user.Password = helpers.CreatePassword(editPage.Password1)
			}
			user.Save()
			url, _ := subRouter.Get("users-list").URL()
			editPage.AddFlashMessage("User Saved")
			http.Redirect(w, r, url.String(), http.StatusFound)
		}
	default:
		{
			http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		}
	}
}
