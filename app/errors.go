package app

import (
	"k.prv/rpimon/app/context"
	l "k.prv/rpimon/logging"
	"net/http"
	"strconv"
)

// Render400 render BadRequest error
func Render400(w http.ResponseWriter, r *http.Request, msgs ...string) {
	msg := "Invalid Request"
	if msgs != nil && len(msgs) > 0 {
		msg = msgs[0]
	}
	ctx := renderError(w, r, http.StatusBadRequest, msg)
	l.Debug("400: user=%s, url=%s", ctx.CurrentUser, r.URL.String())
}

//Render404 render Not Found error page
func Render404(w http.ResponseWriter, r *http.Request, msgs ...string) {
	msg := "Not found"
	if msgs != nil && len(msgs) > 0 {
		msg = msgs[0]
	}
	ctx := renderError(w, r, http.StatusNotFound, msg)
	l.Debug("404: user=%s, url=%s", ctx.CurrentUser, r.URL.String())
}

//Render401 render Unauthorized error page
func Render401(w http.ResponseWriter, r *http.Request, msgs ...string) {
	msg := "Unauthorized"
	if msgs != nil && len(msgs) > 0 {
		msg = msgs[0]
	}
	ctx := renderError(w, r, http.StatusUnauthorized, msg)
	l.Warn("401: user=%s, url=%s", ctx.CurrentUser, r.URL.String())
}

//Render403 render Forbidden error page
func Render403(w http.ResponseWriter, r *http.Request, msgs ...string) {
	msg := "Forbidden"
	if msgs != nil && len(msgs) > 0 {
		msg = msgs[0]
	}
	ctx := renderError(w, r, http.StatusForbidden, msg)
	l.Warn("403: user=%s, url=%s", ctx.CurrentUser, r.URL.String())
}

//Render500 render Internal Server Error page
func Render500(w http.ResponseWriter, r *http.Request, msgs ...string) {
	msg := "Internal Server Error"
	if msgs != nil && len(msgs) > 0 {
		msg = msgs[0]
	}
	ctx := renderError(w, r, http.StatusInternalServerError, msg)
	l.Warn("500: user=%s, url=%s", ctx.CurrentUser, r.URL.String())
}

//RenderError render custom error page
func RenderError(w http.ResponseWriter, r *http.Request, status int, message string) {
	ctx := renderError(w, r, status, message)
	l.Debug("%d: %s; user=%s, url=%s", status, message, ctx.CurrentUser, r.URL.String())
}

type errorContext struct {
	*context.BasePageContext
	Message string
	Status  int
	Error   string
}

func renderError(w http.ResponseWriter, r *http.Request, status int, message string) *errorContext {
	err := "Error " + strconv.Itoa(status)
	ctx := &errorContext{
		BasePageContext: context.NewBasePageContext(err, w, r),
		Message:         message,
		Status:          status,
		Error:           err,
	}
	w.WriteHeader(status)
	RenderTemplateStd(w, ctx, "errors/error.tmpl")
	return ctx
}
