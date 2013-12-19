package app

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/gorilla/context"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"runtime/debug"
	"time"
)

// csrf tokens len
const CSRFTOKENLEN = 64

// csrf tokens name in context
const CONTEXTCSRFTOKEN = "csrf_token"

// csrf tokens name formms
const FORMCSRFTOKEN = "BasePageContext.CsrfToken"

// CSRT Token middleware
func csrfHandler(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := GetSessionStore(w, r)
		csrfToken := session.Get(CONTEXTCSRFTOKEN)
		if csrfToken == nil {
			token := make([]byte, CSRFTOKENLEN)
			rand.Read(token)
			csrfToken = base64.StdEncoding.EncodeToString(token)
			session.Set(CONTEXTCSRFTOKEN, csrfToken)
			session.Save(w, r)
		}

		context.Set(r, CONTEXTCSRFTOKEN, csrfToken)
		if r.Method == "POST" && r.FormValue(FORMCSRFTOKEN) != csrfToken {
			http.Error(w, "Fobidden/CSRF", http.StatusForbidden)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

// EnhResponseWriter response writer with status
type EnhResponseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader store status of request
func (writer *EnhResponseWriter) WriteHeader(status int) {
	writer.ResponseWriter.WriteHeader(status)
	writer.status = status
}

// Logging middleware
func logHandler(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.String()
		method := r.Method
		remote := r.RemoteAddr
		start := time.Now()
		writer := &EnhResponseWriter{ResponseWriter: w, status: 200}

		defer func() {
			end := time.Now()
			status := writer.status
			if err := recover(); err == nil {
				l.Info("%d %s %s %s %s", status, method, url, remote, end.Sub(start))
			} else {
				l.Error("%d %s %s %s %s err:'%#v'", status, method, url, remote, end.Sub(start),
					err)
				l.Error("%v", debug.Stack())
			}
		}()
		h.ServeHTTP(writer, r)
	})
}

// Context middleware
func contextHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer context.Clear(r)
		h.ServeHTTP(w, r)
	})
}
