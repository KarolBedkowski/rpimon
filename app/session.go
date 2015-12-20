package app

import (
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"k.prv/rpimon/cfg"
	l "k.prv/rpimon/logging"
	"net/http"
	"time"
)

const storesession = "SESSION"

// Sessions settings
const (
	sessionLoginKey      = "USERID"
	sessionPermissionKey = "USER_PERM"
	sessionTimestampKey  = "timestamp"
	maxSessionAgeDays    = 5
	maxSessionAge        = time.Duration(24*maxSessionAgeDays) * time.Hour
)

var store *sessions.CookieStore

// initSessionStore initialize sessions support
func initSessionStore(conf *cfg.AppConfiguration) error {
	if len(conf.CookieAuthKey) < 32 {
		l.Info("Random CookieAuthKey")
		conf.CookieAuthKey = string(securecookie.GenerateRandomKey(32))
	}
	if len(conf.CookieEncKey) < 32 {
		l.Info("Random CookieEncKey")
		conf.CookieEncKey = string(securecookie.GenerateRandomKey(32))
	}
	/* for filesystem store
	err := os.MkdirAll(conf.SessionStoreDir, os.ModeDir)
	if err != nil && !os.IsExist(err) {
		l.Error("Createing dir for session store failed ", err)
		return err
	}
	*/
	store = sessions.NewCookieStore([]byte(conf.CookieAuthKey),
		[]byte(conf.CookieEncKey))

	return nil
}

// GetSessionStore  for given request
func GetSessionStore(w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, _ := store.Get(r, storesession)
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * maxSessionAgeDays,
	}
	return session
}

// ClearSession remove all values and save session
func ClearSession(w http.ResponseWriter, r *http.Request) {
	session := GetSessionStore(w, r)
	session.Values = nil
	session.Save(r, w)
}

// SaveSession - shortcut
func SaveSession(w http.ResponseWriter, r *http.Request) error {
	err := sessions.Save(r, w)
	if err != nil {
		l.Error("SaveSession error ", err)
	}
	return err
}

// ClearSessionStore - remove old session files
/*
func ClearSessionStore() error {
	l.Info("Start ClearSessionStore")
	now := time.Now()
	now = now.Add(time.Duration(-24) * time.Hour)
	filepath.Walk(Configuration.SessionStoreDir, func(path string, info os.FileInfo, err error) error {
		if now.After(info.ModTime()) {
			l.Info("Delete ", path)
			os.Remove(path)
		}
		return nil
	})
	return nil
}
*/

// GetLoggerUser return login and permission of logged user
func GetLoggerUser(session *sessions.Session) (login string, permissions []string, ok bool) {
	if slogin := session.Values[sessionLoginKey]; slogin != nil {
		login = slogin.(string)
		ok = true
		if sPerm := session.Values[sessionPermissionKey]; sPerm != nil {
			permissions = sPerm.([]string)
		}
	}
	return
}

// SetLoggedUser save logged user information in session
func SetLoggedUser(s *sessions.Session, login string, privs []string) {
	s.Values[sessionLoginKey] = login
	s.Values[sessionPermissionKey] = privs
	s.Values[sessionTimestampKey] = time.Now().Unix()
}

// SessionHandler check validity of session
func SessionHandler(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		delete(r.Form, "gorilla.csrf.Token")
		s := GetSessionStore(w, r)
		//		context.Set(r, "session", s)
		if ts, ok := s.Values[sessionTimestampKey]; ok {
			timestamp := time.Unix(ts.(int64), 0)
			now := time.Now()
			if now.Sub(timestamp) < maxSessionAge {
				s.Values[sessionTimestampKey] = now.Unix()
			} else {
				// Clear session when expired
				s.Values = nil
			}
			s.Save(r, w)
		}
		//l.Debug("Context: %v", context.GetAll(r))
		h.ServeHTTP(w, r)
	})
}
