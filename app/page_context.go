package app

import (
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"html/template"
	"io/ioutil"
	"k.prv/rpimon/helpers"
	"net/http"
	"strings"
	"time"
)

// BaseCtx context for pages
type BaseCtx struct {
	Session          *sessions.Session
	Title            string
	ResponseWriter   http.ResponseWriter
	Request          *http.Request
	CSRFField        template.HTML
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

// NewBaseCtx create base page context for request
func NewBaseCtx(title string, w http.ResponseWriter, r *http.Request) *BaseCtx {

	s := GetSessionStore(w, r)

	ctx := &BaseCtx{
		Title:          title,
		ResponseWriter: w,
		Request:        r,
		Session:        s,
		CSRFField:      csrf.TemplateField(r),
		Hostname:       hostname,
		Now:            time.Now().Format("2006-01-02 15:04:05"),
		FlashMessages:  make(map[string][]interface{}),
		Version:        AppVersion,
	}
	ctx.CurrentUser, ctx.CurrentUserPerms, _ = GetLoggerUser(s)

	for _, kind := range FlashKind {
		if flashes := s.Flashes(kind); flashes != nil && len(flashes) > 0 {
			ctx.FlashMessages[kind] = flashes
		}
	}

	SetMainMenu(ctx)

	ctx.Save()
	return ctx
}

// GetFlashMessage for current context
func (ctx *BaseCtx) GetFlashMessage() map[string][]interface{} {
	return ctx.FlashMessages
}

// AddFlashMessage to context
func (ctx *BaseCtx) AddFlashMessage(msg interface{}, kind ...string) {
	if len(kind) > 0 {
		ctx.Session.AddFlash(msg, kind...)
	} else {
		ctx.Session.AddFlash(msg, "info")
	}
}

// Set value in session
func (ctx *BaseCtx) Set(key string, value interface{}) {
	ctx.Session.Values[key] = value
}

// Get value from session
func (ctx *BaseCtx) Get(key string) interface{} {
	return ctx.Session.Values[key]
}

// Save session by page context
func (ctx *BaseCtx) Save() error {
	return SaveSession(ctx.ResponseWriter, ctx.Request)
}

// SetMenuActive add id  to menu active items
func (ctx *BaseCtx) SetMenuActive(id string) {
	if ctx.MainMenu == nil {
		return
	}
	ctx.MainMenu.SetActiveMenuItem(id)
}

func (ctx *BaseCtx) Redirect(url string) {
	http.Redirect(ctx.ResponseWriter, ctx.Request, url, http.StatusFound)
}

func (ctx *BaseCtx) RenderStd(context interface{}, args ...string) {
	RenderTemplateStd(ctx.ResponseWriter, context, args...)
}

func (ctx *BaseCtx) Render400(msgs ...string) {
	Render400(ctx.ResponseWriter, ctx.Request, msgs...)
}

// DataPageCtx - context  with data (string) + title
type DataPageCtx struct {
	*BaseCtx
	Data    string
	Header1 string
	Header2 string

	THead []string
	TData [][]string
}

// NewDataPageCtx create new simple context to show text data
func NewDataPageCtx(w http.ResponseWriter, r *http.Request, title string) *DataPageCtx {
	ctx := &DataPageCtx{BaseCtx: NewBaseCtx(title, w, r)}
	return ctx
}

// ContextHandler - handler function called by Context and SecContext
type ContextHandler func(r *http.Request, ctx *BaseCtx)

// Context create BaseCtx for request
func Context(h ContextHandler, title string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewBaseCtx(title, w, r)
		h(r, ctx)
	})
}

// SecContext create BaseCtx for request and check user permissions.
func SecContext(h ContextHandler, title string, permission string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewBaseCtx(title, w, r)
		if ctx.CurrentUser != "" && ctx.CurrentUserPerms != nil {
			if permission == "" {
				h(r, ctx)
				return
			}
			for _, p := range ctx.CurrentUserPerms {
				if p == permission {
					h(r, ctx)
					return
				}
			}
		}
		http.Error(w, "forbidden", http.StatusForbidden)
	})
}
