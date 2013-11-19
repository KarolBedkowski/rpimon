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

type UsersPage struct {
	Users          []database.User
	FlashMesssages []interface{}
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	if GetLoggedUser(w, r, true) == nil {
		return
	}
	a := context.Get(r, "APP").(*app.WebApp)
	data := &UsersPage{
		Users:          database.UsersList(),
		FlashMesssages: a.GetFlashMessage(w, r)}
	a.RenderTemplate(w, "base", data, "base.tmpl", "users/users.tmpl")
}

type UserForm struct {
	database.User
	Password1 string
	Password2 string
	CsrfToken string
}

type EditPage struct {
	Form           *UserForm
	Message        string
	FlashMesssages []interface{}
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
	editPage := &EditPage{
		Form: &UserForm{
			CsrfToken: context.Get(r, app.CONTEXT_CSRF_TOKEN).(string)},
		Message:        "",
		FlashMesssages: a.GetFlashMessage(w, r)}

	switch r.Method {
	case "GET":
		{
			vars := mux.Vars(r)
			userId, ok := vars["id"]
			if ok && userId != "" {
				userIdI, _ := strconv.ParseInt(userId, 10, 64)
				editPage.Form.User = *database.GetUserById(userIdI)
			}
			a.RenderTemplate(w, "base", editPage, "base.tmpl", "users/edit.tmpl")
			return
		}
	case "POST":
		{
			r.ParseForm()
			err := decoder.Decode(editPage.Form, r.Form)
			if err != nil {
				log.Print("Decoding form error", err)
			}
			msg := editPage.Form.validate()
			log.Print("Validate: ", msg)
			if msg != "" {
				editPage.Message = msg
				a.RenderTemplate(w, "base", editPage, "base.tmpl", "users/edit.tmpl")
				return
			}
			var user *database.User
			if editPage.Form.Id > 0 {
				user = database.GetUserById(editPage.Form.Id)
			} else {
				user = new(database.User)
			}
			// Copy
			user.Login = editPage.Form.Login
			user.Name = editPage.Form.Name
			if editPage.Form.Password1 != "" {
				user.Password = helpers.CreatePassword(editPage.Form.Password1)
			}
			user.Save()
			url, _ := subRouter.Get("users-list").URL()
			a.AddFlashMessage(w, r, "User Saved")
			http.Redirect(w, r, url.String(), http.StatusFound)
		}
	default:
		{
			http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		}
	}
}
