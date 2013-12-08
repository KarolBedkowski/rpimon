package app

import (
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"io/ioutil"
	"k.prv/rpimon/helpers"
	"net/http"
	"strings"
)

const STORE_SESSION = "SESSION"

type SessionStore struct {
	Session *sessions.Session
}

func GetSessionStore(w http.ResponseWriter, r *http.Request) *SessionStore {
	session, _ := store.Get(r, STORE_SESSION)
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
	*SessionStore
	Title          string
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	CsrfToken      string
	Hostname       string
	CurrentUser    string
}

var hostname string

func NewBasePageContext(title string, w http.ResponseWriter, r *http.Request) *BasePageContext {
	ctx := &BasePageContext{Title: title,
		ResponseWriter: w,
		Request:        r,
		SessionStore:   GetSessionStore(w, r),
		CsrfToken:      context.Get(r, CONTEXT_CSRF_TOKEN).(string)}
	if hostname == "" {
		file, err := ioutil.ReadFile("/etc/hostname")
		helpers.CheckErr(err, "Load hostname error")
		hostname = strings.Trim(string(file), " \n")
	}
	ctx.Hostname = hostname
	session := GetSessionStore(w, r)
	if session != nil {
		userid := session.Get("USERID")
		if userid != nil {
			ctx.CurrentUser = userid.(string)
		}
	}
	return ctx
}

func (ctx *BasePageContext) GetFlashMessage() []interface{} {
	if flashes := ctx.Session.Flashes(); len(flashes) > 0 {
		err := ctx.SessionSave()
		helpers.CheckErr(err, "GetFlashMessage Save Error")
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
