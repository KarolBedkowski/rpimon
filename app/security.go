package app

import (
	"k.prv/rpimon/database"
	"log"
	"net/http"
)

// Session key for userid
const USERIDSESSION = "USERID"

// GetLoggedUserLogin for request
func GetLoggedUserLogin(w http.ResponseWriter, r *http.Request) (login string) {
	session := GetSessionStore(w, r)
	sessLogin := session.Values[USERIDSESSION]
	if sessLogin != nil {
		login = sessLogin.(string)
	}
	return
}

// GetLoggedUser for request
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

// CheckIsUserLogger for request
func CheckIsUserLogger(w http.ResponseWriter, r *http.Request, redirect bool) (user *database.User) {
	user = GetLoggedUser(w, r)
	if user != nil {
		return
	}
	log.Print("Access denied")
	if redirect {
		url := GetNamedURL("auth-login")
		url += PairsToQuery("back", r.URL.String())
		http.Redirect(w, r, url, 302)
	}
	return
}

// ComparePassword check passwords
func ComparePassword(userPassword string, candidatePassword string) bool {
	return userPassword == candidatePassword
}

// VerifyPermission check is user is logged and have given permission
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

// VerifyLogged check is user is logged
func VerifyLogged(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user := CheckIsUserLogger(w, r, true); user != nil {
			h(w, r)
		}
	})
}
