package app

import (
	"github.com/gorilla/sessions"
	//	"k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const storesession = "SESSION"

// GetSessionStore  for given request
func GetSessionStore(w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, _ := store.Get(r, storesession)
	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 1,
	}
	return session
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
