package app

import (
	"k.prv/rpimon/database"
	"log"
	"net/http"
)

const USERID_SESSION = "USERID"

func GetLoggedUserLogin(w http.ResponseWriter, r *http.Request) (login string) {
	session := GetSessionStore(w, r)
	sessLogin := session.Get(USERID_SESSION)
	if sessLogin != nil {
		login = sessLogin.(string)
	}
	return
}

func GetLoggedUser(w http.ResponseWriter, r *http.Request) (user *database.User) {
	user = nil
	userLogin := GetLoggedUserLogin(w, r)
	if userLogin != "" {
		user := database.GetUserByLogin(userLogin)
		if user != nil {
			return user
		}
	}
	return
}

func CheckIsUserLogger(w http.ResponseWriter, r *http.Request, redirect bool) (user *database.User) {
	user = GetLoggedUser(w, r)
	log.Print("Access denied")
	if redirect {
		url, err := GetNamedUrl("auth-login", "back", r.URL.String())
		if err != nil {
			log.Print("GetNamedUrl error", err)
			return
		}
		http.Redirect(w, r, url, 302)
	}
	return
}
