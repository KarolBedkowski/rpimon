package app

import (
	//"github.com/gorilla/context"
	"crypto/rand"
	"encoding/base64"
	"github.com/gorilla/sessions"
	"io/ioutil"
	"k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"strings"
	"time"
)

// BasePageContext context for pages
type BasePageContext struct {
	Session             *sessions.Session
	Title               string
	ResponseWriter      http.ResponseWriter
	Request             *http.Request
	CsrfToken           string
	Hostname            string
	CurrentUser         string
	MainMenu            []MenuItem
	CurrentMainMenuPos  string
	LocalMenu           []MenuItem
	CurrentLocalMenuPos string
	Now                 string
	FlashMessages       []interface{}
}

var hostname string

// NewBasePageContext create base page context for request
func NewBasePageContext(title string, w http.ResponseWriter, r *http.Request) *BasePageContext {

	session := GetSessionStore(w, r)
	csrfToken := session.Values[CONTEXTCSRFTOKEN]
	if csrfToken == nil {
		l.Print("csrfToken is nil - creating")
		token := make([]byte, CSRFTOKENLEN)
		rand.Read(token)
		csrfToken = base64.StdEncoding.EncodeToString(token)
		session.Values[CONTEXTCSRFTOKEN] = csrfToken
		session.Save(r, w)
	}

	ctx := &BasePageContext{Title: title,
		ResponseWriter: w,
		Request:        r,
		Session:        session,
		CsrfToken:      csrfToken.(string)}
	if hostname == "" {
		file, err := ioutil.ReadFile("/etc/hostname")
		helpers.CheckErr(err, "Load hostname error")
		hostname = strings.Trim(string(file), " \n")
	}
	ctx.Hostname = hostname
	ctx.CurrentUser = GetLoggedUserLogin(w, r)
	ctx.Now = time.Now().Format("2006-01-02 15:04:05")
	SetMainMenu(ctx, ctx.CurrentUser != "")

	if flashes := ctx.Session.Flashes(); len(flashes) > 0 {
		ctx.FlashMessages = flashes
		ctx.Save()
	}
	return ctx
}

// GetFlashMessage for current context
func (ctx *BasePageContext) GetFlashMessage() []interface{} {
	return ctx.FlashMessages
}

// AddFlashMessage to context
func (ctx *BasePageContext) AddFlashMessage(msg interface{}) {
	ctx.Session.AddFlash(msg)
}

func (ctx *BasePageContext) Set(key string, value interface{}) {
	ctx.Session.Values[key] = value
}

// SessionSave by page context
func (ctx *BasePageContext) Save() error {
	err := sessions.Save(ctx.Request, ctx.ResponseWriter)
	helpers.CheckErr(err, "BasePageContext Save Error")
	return err
}
