package app

import (
	"crypto/md5"
	"fmt"
	"github.com/gorilla/context"
	"io"
	"k.prv/rpimon/database"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"log"
	"net/http"
	"time"
)

// Session key for userid
const sessionLoginKey = "USERID"
const usercontextkey = "USER"

const sessionTimestampKey = "timestamp"

const maxSessionAge = time.Duration(24) * time.Hour

// GetLoggedUserLogin for request
func GetLoggedUserLogin(w http.ResponseWriter, r *http.Request) (login string) {
	session := GetSessionStore(w, r)
	if ts, ok := session.Values[sessionTimestampKey]; ok {
		timestamp := time.Unix(ts.(int64), 0)
		now := time.Now()
		if now.Sub(timestamp) < maxSessionAge {
			session.Values[sessionTimestampKey] = now.Unix()
			session.Save(r, w)
			sessLogin := session.Values[sessionLoginKey]
			if sessLogin != nil {
				login = sessLogin.(string)
			}
		}
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

// GetUser from database based on login
func GetUser(login string) (user *database.User) {
	if login != "" {
		user := database.GetUserByLogin(login)
		if user != nil {
			return user
		}
	}
	return nil
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
		url += h.BuildQuery("back", r.URL.String())
		http.Redirect(w, r, url, 302)
	}
	return
}

// ComparePassword check passwords
func ComparePassword(userPassword string, candidatePassword string) bool {
	hash := md5.New()
	io.WriteString(hash, candidatePassword)
	pass := fmt.Sprintf("%x", hash.Sum(nil))
	return pass == userPassword
}

// VerifyPermission check is user is logged and have given permission
func VerifyPermission(h http.HandlerFunc, permission string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user := CheckIsUserLogger(w, r, true); user != nil {
			if user.HasPermission(permission) {
				context.Set(r, usercontextkey, user)
				h(w, r)
				return
			}
			http.Error(w, "Fobidden/Privilages", http.StatusForbidden)
		}
	})
}

// VerifyLogged check is user is logged
func VerifyLogged(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user := CheckIsUserLogger(w, r, true); user != nil {
			context.Set(r, usercontextkey, user)
			h(w, r)
		}
	})
}

// LoginUser - update session
func LoginUser(w http.ResponseWriter, r *http.Request, login string) error {
	l.Info("User %s log in", login)
	session := GetSessionStore(w, r)
	session.Values[sessionLoginKey] = login
	session.Values[sessionTimestampKey] = time.Now().Unix()
	session.Save(r, w)
	return nil
}
