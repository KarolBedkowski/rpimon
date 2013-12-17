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

const STORE_SESSION = "SESSION"

type SessionStore struct {
	Session *sessions.Session
}

func GetSessionStore(w http.ResponseWriter, r *http.Request) *SessionStore {
	session, _ := store.Get(r, STORE_SESSION)
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 1,
	}
	return &SessionStore{session}
}

func (store *SessionStore) Get(key string) interface{} {
	return store.Session.Values[key]
}

func (store *SessionStore) Set(key string, value interface{}) {
	store.Session.Values[key] = value
}

func (store *SessionStore) Clear() {
	store.Session.Values = nil
}

func (store *SessionStore) Save(w http.ResponseWriter, r *http.Request) error {
	err := store.Session.Save(r, w)
	helpers.CheckErr(err, "BasePageContext Save Error")
	return err
}

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
