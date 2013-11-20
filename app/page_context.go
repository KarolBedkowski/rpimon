package app

import (
	"../helpers"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

const STORE_SESSION = "SESSION"

type SessionStore struct {
	Session *sessions.Session
}

func GetSessionStore(w http.ResponseWriter, r *http.Request) *SessionStore {
	session, _ := App.store.Get(r, STORE_SESSION)
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

type BasePageContext struct {
	Title          string
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	*SessionStore
	CsrfToken string
}

func NewBasePageContext(title string, w http.ResponseWriter, r *http.Request) *BasePageContext {
	ctx := &BasePageContext{title, w, r, GetSessionStore(w, r), ""}
	return ctx
}

func (ctx *BasePageContext) GetFlashMessage() []interface{} {
	if flashes := ctx.Session.Flashes(); len(flashes) > 0 {
		err := ctx.SessionSave()
		helpers.CheckErr(err, "GetFlashMessage Save Error")
		log.Print("GetFlashMessage ", flashes, ctx.Session.Flashes())
		return flashes
	}
	return nil
}

func (ctx *BasePageContext) AddFlashMessage(msg interface{}) {
	ctx.Session.AddFlash(msg)
}

func (ctx *BasePageContext) SessionSave() error {
	err := ctx.Session.Save(ctx.Request, ctx.ResponseWriter)
	helpers.CheckErr(err, "BasePageContext Save Error")
	return err
}
