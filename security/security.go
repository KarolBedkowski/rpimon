package security

import (
	"../app"
	"../database"
	"log"
	"net/http"
)

type Credentials struct {
	User *database.User
}

const USERID_SESSION = "userid"
const USERLOGIN_SESSION = "userlogin"

func GetLoggedUser(w http.ResponseWriter, r *http.Request, redirect bool) (credentials *Credentials) {
	credentials = nil
	session := app.GetSessionStore(w, r)
	userId := session.Get(USERID_SESSION)
	if userId != nil {
		user := database.GetUserById(userId.(int64))
		if user != nil {
			credentials = &Credentials{User: user}
			return
		}
	}
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
