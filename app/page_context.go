package app

import (
	"github.com/gorilla/context"
	"io/ioutil"
	"k.prv/rpimon/helpers"
	"net/http"
	"strings"
	"time"
)

// BasePageContext context for pages
type BasePageContext struct {
	*sessionStore
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
}

var hostname string

// NewBasePageContext create base page context for request
func NewBasePageContext(title string, w http.ResponseWriter, r *http.Request) *BasePageContext {
	ctx := &BasePageContext{Title: title,
		ResponseWriter: w,
		Request:        r,
		sessionStore:   GetSessionStore(w, r),
		CsrfToken:      context.Get(r, CONTEXTCSRFTOKEN).(string)}
	if hostname == "" {
		file, err := ioutil.ReadFile("/etc/hostname")
		helpers.CheckErr(err, "Load hostname error")
		hostname = strings.Trim(string(file), " \n")
	}
	ctx.Hostname = hostname
	ctx.CurrentUser = GetLoggedUserLogin(w, r)
	ctx.Now = time.Now().Format("2006-01-02 15:04:05")
	SetMainMenu(ctx, ctx.CurrentUser != "")
	return ctx
}

// GetFlashMessage for current context
func (ctx *BasePageContext) GetFlashMessage() []interface{} {
	if flashes := ctx.Session.Flashes(); len(flashes) > 0 {
		err := ctx.SessionSave()
		helpers.CheckErr(err, "GetFlashMessage Save Error")
		return flashes
	}
	return nil
}

// AddFlashMessage to context
func (ctx *BasePageContext) AddFlashMessage(msg interface{}) {
	ctx.Session.AddFlash(msg)
}

// SessionSave by page context
func (ctx *BasePageContext) SessionSave() error {
	err := ctx.Session.Save(ctx.Request, ctx.ResponseWriter)
	helpers.CheckErr(err, "BasePageContext Save Error")
	return err
}
