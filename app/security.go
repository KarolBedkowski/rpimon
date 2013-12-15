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
	if user != nil {
		return
	}
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

func ComparePassword(user_password string, candidate_password string) bool {
	return user_password == candidate_password
}

func VerifyPermission(h http.HandlerFunc, permission string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user := CheckIsUserLogger(w, r, true); user != nil {
			for _, perm := range user.Privs {
				if perm == permission {
					h(w, r)
					return
				}
			}
			http.Error(w, "Fobidden/Privilages", http.StatusForbidden)
		}
	})
}

func VerifyLogged(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user := CheckIsUserLogger(w, r, true); user != nil {
			h(w, r)
		}
	})
}
