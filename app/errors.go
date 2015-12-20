package app

import (
	l "k.prv/rpimon/logging"
	"net/http"
	"strconv"
)

// Render400 render BadRequest error
func Render400(w http.ResponseWriter, r *http.Request, msgs ...string) {
	ctx := renderError(w, r, http.StatusBadRequest, msgs...)
	l.Debug("400: user=%s, url=%s", ctx.CurrentUser, r.URL.String())
}

//Render404 render Not Found error page
func Render404(w http.ResponseWriter, r *http.Request, msgs ...string) {
	ctx := renderError(w, r, http.StatusNotFound, msgs...)
	l.Debug("404: user=%s, url=%s", ctx.CurrentUser, r.URL.String())
}

//Render401 render Unauthorized error page
func Render401(w http.ResponseWriter, r *http.Request, msgs ...string) {
	ctx := renderError(w, r, http.StatusUnauthorized, msgs...)
	l.Warn("401: user=%s, url=%s", ctx.CurrentUser, r.URL.String())
}

//Render403 render Forbidden error page
func Render403(w http.ResponseWriter, r *http.Request, msgs ...string) {
	ctx := renderError(w, r, http.StatusForbidden, msgs...)
	l.Warn("403: user=%s, url=%s", ctx.CurrentUser, r.URL.String())
}

//Render500 render Internal Server Error page
func Render500(w http.ResponseWriter, r *http.Request, msgs ...string) {
	ctx := renderError(w, r, http.StatusInternalServerError, msgs...)
	l.Warn("500: user=%s, url=%s", ctx.CurrentUser, r.URL.String())
}

//RenderError render custom error page
func RenderError(w http.ResponseWriter, r *http.Request, status int, message string) {
	ctx := renderError(w, r, status, message)
	l.Debug("%d: %s; user=%s, url=%s", status, message, ctx.CurrentUser, r.URL.String())
}

type errorContext struct {
	*BaseCtx
	Message string
	Status  int
	Error   string
}

func renderError(w http.ResponseWriter, r *http.Request, status int, messages ...string) *errorContext {
	err := "Error " + strconv.Itoa(status)
	ctx := &errorContext{
		BaseCtx: NewBaseCtx(err, w, r),
		Status:  status,
		Error:   err,
	}
	if messages != nil && len(messages) > 0 {
		ctx.Message = messages[0]
	} else {
		ctx.Message = http.StatusText(status)
	}
	w.WriteHeader(status)
	RenderTemplateStd(w, ctx, "errors/error.tmpl")
	return ctx
}
