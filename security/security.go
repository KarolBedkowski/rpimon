package security

import (
	"k.prv/rpimon/app"
	"k.prv/rpimon/database"
	"log"
	"net/http"
)

const USERID_SESSION = "USERID"

func GetLoggedUser(w http.ResponseWriter, r *http.Request) (user *database.User) {
	user = nil
	session := app.GetSessionStore(w, r)
	userLogin := session.Get(USERID_SESSION)
	if userLogin != nil {
		user := database.GetUserByLogin(userLogin.(string))
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
		url, err := app.GetNamedUrl("auth-login", "back", r.URL.String())
		if err != nil {
			log.Print("GetNamedUrl error", err)
			return
		}
		http.Redirect(w, r, url, 302)
	}
	return
}
