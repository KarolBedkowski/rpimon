package users

import (
	"../../app"
	"../../database"
	"../../helpers"
	"../../security"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"log"
	"net/http"
	"strconv"
)

var subRouter *mux.Router
var decoder = schema.NewDecoder()

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
	if security.GetLoggedUser(w, r, true) == nil {
		return
	}
	data := newUsersPageCtx(w, r, database.UsersList())
	app.RenderTemplate(w, data, "base", "base.tmpl", "users/users.tmpl", "flash.tmpl")
}

type UserForm struct {
	database.User
	Password1 string
	Password2 string
}

type EditPageCtx struct {
	*app.BasePageContext
	*UserForm
	Message string
}

func newEditPageCtx(w http.ResponseWriter, r *http.Request, msg string) *EditPageCtx {
	return &EditPageCtx{app.NewBasePageContext("User", w, r),
		&UserForm{}, msg}
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
	if security.GetLoggedUser(w, r, true) == nil {
		return
	}
	editPage := newEditPageCtx(w, r, "")
	switch r.Method {
	case "GET":
		{
			vars := mux.Vars(r)
			userId, ok := vars["id"]
			if ok && userId != "" {
				userIdI, _ := strconv.ParseInt(userId, 10, 64)
				editPage.User = *database.GetUserById(userIdI)
			}
			app.RenderTemplate(w, editPage, "base", "base.tmpl", "users/edit.tmpl", "flash.tmpl")
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
				app.RenderTemplate(w, editPage, "base", "base.tmpl", "users/edit.tmpl", "flash.tmpl")
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
			editPage.SessionSave()
			http.Redirect(w, r, url.String(), http.StatusFound)
		}
	default:
		{
			http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		}
	}
}
