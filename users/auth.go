package users

import (
	"../app"
	"../database"
	"../helpers"
	"github.com/gorilla/context"
	"github.com/gorilla/schema"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const PROFILE_SESSION = "profile"
const USERID_SESSION = "userid"
const USERLOGIN_SESSION = "userid"

var decoder = schema.NewDecoder()

type Credentials struct {
	User *database.User
}

func GetLoggedUser(w http.ResponseWriter, r *http.Request, redirect bool) (credentials *Credentials) {
	credentials = nil
	a := context.Get(r, "APP").(*app.WebApp)
	session := app.NewSessionStore(w, r)
	userId := session.Get(USERID_SESSION)
	if userId != nil {
		userIdI, _ := strconv.ParseInt(userId.(string), 10, 64)
		user := database.GetUserById(userIdI)
		if user != nil {
			credentials = &Credentials{User: user}
			return
		}
	}
	log.Print("Access denied")
	if redirect {
		login_url, _ := a.Router.Get("auth-login").URL()
		durl := login_url.String() + "?back=" + url.QueryEscape(r.URL.String())
		http.Redirect(w, r, durl, 302)
	}
	return
}

type LoginPage struct {
	FlashMesssages []interface{}
	Login          string
	Password       string
	Message        string
	back           string
	CsrfToken      string
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	a := context.Get(r, "APP").(*app.WebApp)
	loginPage := &LoginPage{
		Message:        "",
		FlashMesssages: a.GetFlashMessage(w, r),
		Login:          "",
		Password:       "",
		CsrfToken:      context.Get(r, app.CONTEXT_CSRF_TOKEN).(string)}

	session := app.NewSessionStore(w, r)

	switch r.Method {
	case "GET":
		{
			a.RenderTemplate(w, "base", loginPage, "login.tmpl")
			return
		}
	case "POST":
		{
			_ = r.ParseForm()
			values := r.Form
			err := decoder.Decode(loginPage, values)
			if err != nil {
				log.Print("Decode form error", err, values)
			}
			password := loginPage.Password
			if password == "" || loginPage.Login == "" {
				loginPage.Message = "Missing login and/or password"
				a.RenderTemplate(w, "base", loginPage, "base.tmpl", "login.tmpl")
				return
			}
			user := database.GetUserByLogin(loginPage.Login)
			if user != nil {
				cp_err := helpers.ComparePassword(user.Password, password)
				if cp_err != nil {
					loginPage.Message = "Wrong user or password"
					a.RenderTemplate(w, "base", loginPage, "login.tmpl")
					return
				}
				log.Printf("User %s log in", user.Login)
			}
			session.Set(USERID_SESSION, user.Id)
			session.Set(USERLOGIN_SESSION, user.Login)
			session.Save()
			log.Print("values", values, loginPage.back)
			if values["back"] != nil && values["back"][0] != "" {
				log.Print("Red", values["back"][0])
				http.Redirect(w, r, values["back"][0], http.StatusFound)
			} else {
				http.Redirect(w, r, "/", http.StatusFound)
			}
		}

	}
}

func LogoffHandler(w http.ResponseWriter, r *http.Request) {
	session := app.NewSessionStore(w, r)
	session.Clear()
	session.Save()
	http.Redirect(w, r, "/", http.StatusFound)

}
