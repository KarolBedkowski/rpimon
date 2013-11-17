package users

import (
	"../app"
	"../database"
	"../helpers"
	_ "github.com/gorilla/mux"
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
	session, err := app.App.CookieStore.Get(r, PROFILE_SESSION)
	helpers.CheckErr(err, "missing session")
	userId, ok := session.Values[USERID_SESSION].(string)
	if ok && userId != "" {
		userIdI, _ := strconv.ParseInt(userId, 10, 64)
		user := database.GetUserById(userIdI)
		if user != nil {
			credentials = &Credentials{User: user}
			return
		}
	}
	log.Print("Access denied")
	if redirect {
		log.Print("back=", r.URL.Path)
		login_url, _ := app.App.Router.Get("auth-login").URL()
		durl := login_url.String() + "?back=" + url.QueryEscape(r.URL.String())
		log.Print("red=", durl)
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
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	//	vars := mux.Vars(r)
	loginPage := &LoginPage{
		Message:        "",
		FlashMesssages: app.GetFlashMessage(w, r),
		Login:          "",
		Password:       ""}

	switch r.Method {
	case "GET":
		{
			session, _ := app.App.CookieStore.Get(r, PROFILE_SESSION)
			session.Values[USERID_SESSION] = nil
			session.Values[USERLOGIN_SESSION] = nil
			session.Save(r, w)
			app.App.RenderTemplate(w, "base", loginPage, "base.tmpl", "login.tmpl")
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
				app.App.RenderTemplate(w, "base", loginPage, "base.tmpl", "login.tmpl")
				return
			}
			user := database.GetUserByLogin(loginPage.Login)
			if user != nil {
				cp_err := helpers.ComparePassword(user.Password, password)
				if cp_err != nil {
					loginPage.Message = "Wrong user or password"
					app.App.RenderTemplate(w, "base", loginPage, "base.tmpl", "login.tmpl")
					return
				}
				log.Printf("User %s log in", user.Login)
				app.AddFlashMessage(w, r, "User Log in..")
			}
			session, _ := app.App.CookieStore.Get(r, PROFILE_SESSION)
			session.Values[USERID_SESSION] = user.Id
			session.Values[USERLOGIN_SESSION] = user.Login
			session.Save(r, w)
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
	session, _ := app.App.CookieStore.Get(r, PROFILE_SESSION)
	session.Values[USERID_SESSION] = nil
	session.Values[USERLOGIN_SESSION] = nil
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)

}
