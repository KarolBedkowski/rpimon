package app

import (
	"github.com/gorilla/sessions"
	"k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const storesession = "SESSION"

type sessionStore struct {
	Session *sessions.Session
}

// GetSessionStore  for given request
func GetSessionStore(w http.ResponseWriter, r *http.Request) *sessionStore {
	session, _ := store.Get(r, storesession)
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 1,
	}
	return &sessionStore{session}
}

// Get value from session store
func (store *sessionStore) Get(key string) interface{} {
	return store.Session.Values[key]
}

// Set value in session store
func (store *sessionStore) Set(key string, value interface{}) {
	store.Session.Values[key] = value
}

// Clear session
func (store *sessionStore) Clear() {
	store.Session.Values = nil
}

// Save session
func (store *sessionStore) Save(w http.ResponseWriter, r *http.Request) error {
	err := store.Session.Save(r, w)
	helpers.CheckErr(err, "BasePageContext Save Error")
	return err
}

// ClearSessionStore - remove old session files
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
