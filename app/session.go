package app

import (
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	//	"k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	//	"os"
	//	"path/filepath"
	//	"time"
)

const storesession = "SESSION"

var store *sessions.CookieStore

func initSessionStore(conf *AppConfiguration) error {
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
		MaxAge: 86400 * 1,
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
