package app

import (
	"k.prv/rpimon/app/cfg"
	"k.prv/rpimon/app/session"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"log"
	"net/http"
	"time"
)

// Session key
const (
	sessionLoginKey      = "USERID"
	sessionPermissionKey = "USER_PERM"
)

// Sessions settings
const (
	sessionTimestampKey = "timestamp"
	maxSessionAge       = time.Duration(24) * time.Hour
)

// GetLoggedUserInfo returns current login and permission
func GetLoggedUserInfo(w http.ResponseWriter, r *http.Request) (login string, perm []string) {
	s := session.GetSessionStore(w, r)
	if ts, ok := s.Values[sessionTimestampKey]; ok {
		timestamp := time.Unix(ts.(int64), 0)
		now := time.Now()
		if now.Sub(timestamp) < maxSessionAge {
			s.Values[sessionTimestampKey] = now.Unix()
			s.Save(r, w)
			if sessLogin := s.Values[sessionLoginKey]; sessLogin != nil {
				login = sessLogin.(string)
				if sessPerm := s.Values[sessionPermissionKey]; sessPerm != nil {
					perm = sessPerm.([]string)
				}
			}
		}
	}
	return
}

// CheckUserLoggerOrRedirect for request; if user is not logged - redirect to login page
func CheckUserLoggerOrRedirect(w http.ResponseWriter, r *http.Request) (login string, perm []string) {
	login, perm = GetLoggedUserInfo(w, r)
	if login != "" {
		return
	}
	log.Print("Access denied")
	url := GetNamedURL("auth-login")
	url += h.BuildQuery("back", r.URL.String())
	http.Redirect(w, r, url, 302)
	return
}

// VerifyPermission check is user is logged and have given permission
func VerifyPermission(h http.HandlerFunc, permission string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if login, perms := CheckUserLoggerOrRedirect(w, r); login != "" {
			if CheckPermission(perms, permission) {
				h(w, r)
				return
			}
			http.Error(w, "Fobidden/Privilages", http.StatusForbidden)
		}
	})
}

// VerifyLogged check only is user is logged
func VerifyLogged(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if login, _ := CheckUserLoggerOrRedirect(w, r); login != "" {
			h(w, r)
		}
	})
}

// LoginUser - update session
func LoginUser(w http.ResponseWriter, r *http.Request, user *cfg.User) error {
	l.Info("User %s log in", user.Login)
	s := session.GetSessionStore(w, r)
	s.Values[sessionLoginKey] = user.Login
	s.Values[sessionPermissionKey] = user.Privs
	s.Values[sessionTimestampKey] = time.Now().Unix()
	s.Save(r, w)
	return nil
}

// CheckPermission return true if requred permission is on list
func CheckPermission(userPermissions []string, required string) bool {
	if required == "" {
		return true
	}
	for _, p := range userPermissions {
		if p == required {
			return true
		}
	}
	return false
}
