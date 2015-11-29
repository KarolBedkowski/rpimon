package context

import (
	"github.com/gorilla/sessions"
	"io/ioutil"
	"k.prv/rpimon/app/mw"
	asess "k.prv/rpimon/app/session"
	"k.prv/rpimon/helpers"
	//	l "k.prv/rpimon/logging"
	"net/http"
	"strings"
	"time"
)

// BasePageContext context for pages
type BasePageContext struct {
	Session          *sessions.Session
	Title            string
	ResponseWriter   http.ResponseWriter
	Request          *http.Request
	CsrfToken        string
	Hostname         string
	CurrentUser      string
	CurrentUserPerms []string
	MainMenu         *MenuItem
	Now              string
	FlashMessages    map[string][]interface{}
	Tabs             []*MenuItem
	Version          string
}

var hostname string

// AppVersion holds application build version
var AppVersion = "dev"

func init() {
	file, err := ioutil.ReadFile("/etc/hostname")
	helpers.CheckErr(err, "Load hostname error")
	hostname = strings.Trim(string(file), " \n")
}

// Types of flashes
var FlashKind = []string{"error", "info", "success"}

// NewBasePageContext create base page context for request
func NewBasePageContext(title string, w http.ResponseWriter, r *http.Request) *BasePageContext {

	s := asess.GetSessionStore(w, r)
	csrfToken := s.Values[mw.CONTEXTCSRFTOKEN]
	if csrfToken == nil {
		csrfToken = mw.CreateNewCsrfToken()
		s.Values[mw.CONTEXTCSRFTOKEN] = csrfToken
	}

	ctx := &BasePageContext{Title: title,
		ResponseWriter: w,
		Request:        r,
		Session:        s,
		CsrfToken:      csrfToken.(string),
		Hostname:       hostname,
		Now:            time.Now().Format("2006-01-02 15:04:05"),
		FlashMessages:  make(map[string][]interface{}),
		Version:        AppVersion,
	}
	ctx.CurrentUser, ctx.CurrentUserPerms, _ = asess.GetLoggerUser(s)

	for _, kind := range FlashKind {
		if flashes := ctx.Session.Flashes(kind); flashes != nil && len(flashes) > 0 {
			ctx.FlashMessages[kind] = flashes
		}
	}

	SetMainMenu(ctx)

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
	return asess.SaveSession(ctx.ResponseWriter, ctx.Request)
}

// SetMenuActive add id  to menu active items
func (ctx *BasePageContext) SetMenuActive(id string) {
	if ctx.MainMenu == nil {
		return
	}
	ctx.MainMenu.SetActiveMenuItem(id)
}

// SimpleDataPageCtx - context  with data (string) + title
type SimpleDataPageCtx struct {
	*BasePageContext
	Data    string
	Header1 string
	Header2 string

	THead []string
	TData [][]string
}

// NewSimpleDataPageCtx create new simple context to show text data
func NewSimpleDataPageCtx(w http.ResponseWriter, r *http.Request, title string) *SimpleDataPageCtx {
	ctx := &SimpleDataPageCtx{BasePageContext: NewBasePageContext(title, w, r)}
	return ctx
}

// ContextHandler - handler function called by Context and SecContext
type ContextHandler func(w http.ResponseWriter, r *http.Request, ctx *BasePageContext)

// Context create BasePageContext for request
func Context(h ContextHandler, title string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewBasePageContext(title, w, r)
		h(w, r, ctx)
	})
}

// SecContext create BasePageContext for request and check user permissions.
func SecContext(h ContextHandler, title string, permission string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewBasePageContext(title, w, r)
		if ctx.CurrentUser != "" && ctx.CurrentUserPerms != nil {
			if permission == "" {
				h(w, r, ctx)
				return
			}
			for _, p := range ctx.CurrentUserPerms {
				if p == permission {
					h(w, r, ctx)
					return
				}
			}
		}
		http.Error(w, "forbidden", http.StatusForbidden)
	})
}
