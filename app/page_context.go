package app

import (
	"github.com/gorilla/sessions"
	"io/ioutil"
	"k.prv/rpimon/helpers"
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
	LocalMenu           []MenuItem
	CurrentMainMenuPos  string
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
		csrfToken = createNewCsrfToken()
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
