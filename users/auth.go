package users

import (
	"../app"
	"../database"
	"../helpers"
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
	session := app.GetSessionStore(w, r)
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
		login_url, _ := app.Router.Get("auth-login").URL()
		durl := login_url.String() + "?back=" + url.QueryEscape(r.URL.String())
		http.Redirect(w, r, durl, 302)
	}
	return
}

type LoginForm struct {
	Login    string
	Password string
	Message  string
}

type LoginPageCtx struct {
	*app.BasePageContext
	*LoginForm
	back string
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	loginPageCtx := &LoginPageCtx{app.NewBasePageContext("Login", w, r),
		new(LoginForm), ""}
	switch r.Method {
	case "GET":
		{
			app.RenderTemplate(w, loginPageCtx, "base", "login.tmpl", "flash.tmpl")
			return
		}
	case "POST":
		{
			_ = r.ParseForm()
			values := r.Form
			err := decoder.Decode(loginPageCtx, values)
			if err != nil {
				log.Print("Decode form error", err, values)
			}
			password := loginPageCtx.Password
			if password == "" || loginPageCtx.Login == "" {
				loginPageCtx.Message = "Missing login and/or password"
				app.RenderTemplate(w, loginPageCtx, "base", "base.tmpl", "login.tmpl", "flash.tmpl")
				return
			}
			user := database.GetUserByLogin(loginPageCtx.Login)
			if user != nil {
				cp_err := helpers.ComparePassword(user.Password, password)
				if cp_err != nil {
					loginPageCtx.Message = "Wrong user or password"
					app.RenderTemplate(w, loginPageCtx, "base", "login.tmpl", "flash.tmpl")
					return
				}
				log.Printf("User %s log in", user.Login)
			}
			loginPageCtx.Set(USERID_SESSION, user.Id)
			loginPageCtx.Set(USERLOGIN_SESSION, user.Login)
			loginPageCtx.SessionSave()
			if values["back"] != nil && values["back"][0] != "" {
				log.Print("Redirect to ", values["back"][0])
				http.Redirect(w, r, values["back"][0], http.StatusFound)
			} else {
				http.Redirect(w, r, "/", http.StatusFound)
			}
		}
	}
}

func LogoffHandler(w http.ResponseWriter, r *http.Request) {
	session := app.GetSessionStore(w, r)
	session.Clear()
	session.Save(w, r)
	http.Redirect(w, r, "/", http.StatusFound)

}
