package app

import (
	"github.com/gorilla/context"
	"io/ioutil"
	"k.prv/rpimon/helpers"
	"net/http"
	"strings"
)

type BasePageContext struct {
	*SessionStore
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
	ctx.CurrentUser = GetLoggedUserLogin(w, r)
	SetMainMenu(ctx, ctx.CurrentUser != "")
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
