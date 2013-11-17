package users

import (
	"../app"
	"../database"
	"../helpers"
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
	data := &UsersPage{
		Users:          database.UsersList(),
		FlashMesssages: app.GetFlashMessage(w, r)}
	app.App.RenderTemplate(w, "base", data, "base.tmpl", "users/users.tmpl")
}

type UserForm struct {
	database.User
	Password1 string
	Password2 string
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
	editPage := &EditPage{
		Form:           new(UserForm),
		Message:        "",
		FlashMesssages: app.GetFlashMessage(w, r)}

	switch r.Method {
	case "GET":
		{
			vars := mux.Vars(r)
			userId := vars["id"]
			if userId != "" {
				userIdI, _ := strconv.ParseInt(userId, 10, 64)
				editPage.Form.User = *database.GetUserById(userIdI)
			}
			app.App.RenderTemplate(w, "base", editPage, "base.tmpl", "users/edit.tmpl")
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
				app.App.RenderTemplate(w, "base", editPage, "base.tmpl", "users/edit.tmpl")
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
			app.AddFlashMessage(w, r, "User Saved")
			http.Redirect(w, r, url.String(), http.StatusFound)
		}
	default:
		{
			http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		}
	}
}
