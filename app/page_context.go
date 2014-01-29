package app

import (
	"github.com/gorilla/sessions"
	"io/ioutil"
	"k.prv/rpimon/helpers"
	//	l "k.prv/rpimon/helpers/logging"
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
	MainMenu            []*MenuItem
	LocalMenu           []*MenuItem
	CurrentMainMenuPos  string
	CurrentLocalMenuPos string
	Now                 string
	FlashMessages       map[string][]interface{}
}

var hostname string

func init() {
	file, err := ioutil.ReadFile("/etc/hostname")
	helpers.CheckErr(err, "Load hostname error")
	hostname = strings.Trim(string(file), " \n")
}

// Types of flashes
var FlashKind = []string{"error", "info", "success"}

// NewBasePageContext create base page context for request
func NewBasePageContext(title, mainMenuID string, w http.ResponseWriter, r *http.Request) *BasePageContext {

	session := GetSessionStore(w, r)
	csrfToken := session.Values[CONTEXTCSRFTOKEN]
	if csrfToken == nil {
		csrfToken = createNewCsrfToken()
		session.Values[CONTEXTCSRFTOKEN] = csrfToken
		session.Save(r, w)
	}

	ctx := &BasePageContext{Title: title,
		ResponseWriter:     w,
		Request:            r,
		Session:            session,
		CsrfToken:          csrfToken.(string),
		Hostname:           hostname,
		CurrentUser:        GetLoggedUserLogin(w, r),
		Now:                time.Now().Format("2006-01-02 15:04:05"),
		CurrentMainMenuPos: mainMenuID,
		FlashMessages:      make(map[string][]interface{}),
	}

	SetMainMenu(ctx)

	for _, kind := range FlashKind {
		if flashes := ctx.Session.Flashes(kind); flashes != nil && len(flashes) > 0 {
			ctx.FlashMessages[kind] = flashes
		}
	}
	ctx.Save()
	return ctx
}

// GetFlashMessage for current context
func (ctx *BasePageContext) GetFlashMessage() map[string][]interface{} {
	return ctx.FlashMessages
}

// AddFlashMessage to context
func (ctx *BasePageContext) AddFlashMessage(msg interface{}, kind ...string) {
	if len(kind) > 0 {
		ctx.Session.AddFlash(msg, kind...)
	} else {
		ctx.Session.AddFlash(msg, "info")
	}
}

// Set value in session
func (ctx *BasePageContext) Set(key string, value interface{}) {
	ctx.Session.Values[key] = value
}

// Get value from session
func (ctx *BasePageContext) Get(key string) interface{} {
	return ctx.Session.Values[key]
}

// Save session by page context
func (ctx *BasePageContext) Save() error {
	return SaveSession(ctx.ResponseWriter, ctx.Request)
}

// SimpleDataPageCtx - context  with data (string) + title
type SimpleDataPageCtx struct {
	*BasePageContext
	CurrentPage string
	Data        string

	THead []string
	TData [][]string
}

// NewSimpleDataPageCtx create new simple context to show text data
func NewSimpleDataPageCtx(w http.ResponseWriter, r *http.Request,
	title string, mainMenuID string, cuurentPage string, localMenu []*MenuItem) *SimpleDataPageCtx {
	ctx := &SimpleDataPageCtx{BasePageContext: NewBasePageContext(title, mainMenuID, w, r)}
	ctx.LocalMenu = localMenu
	return ctx
}
