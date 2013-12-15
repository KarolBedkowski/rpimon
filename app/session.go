package app

import (
	"github.com/gorilla/sessions"
	"k.prv/rpimon/helpers"
	"net/http"
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
