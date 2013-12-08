package app

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/gorilla/context"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"time"
)

// csrf tokens
const CSRF_TOKEN_LEN = 64
const CONTEXT_CSRF_TOKEN = "csrf_token"
const FORM_CSRF_TOKEN = "BasePageContext.CsrfToken"

// CSRT Token middleware
func csrfHandler(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := GetSessionStore(w, r)
		csrf_token := session.Get(CONTEXT_CSRF_TOKEN)
		if csrf_token == nil {
			token := make([]byte, CSRF_TOKEN_LEN)
			rand.Read(token)
			csrf_token = base64.StdEncoding.EncodeToString(token)
			session.Set(CONTEXT_CSRF_TOKEN, csrf_token)
			session.Save(w, r)
		}

		context.Set(r, CONTEXT_CSRF_TOKEN, csrf_token)
		if r.Method == "POST" && r.FormValue(FORM_CSRF_TOKEN) != csrf_token {
			http.Error(w, "Fobidden/CSRF", http.StatusForbidden)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

type EnhResponseWriter struct {
	http.ResponseWriter
	status int
}

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
			err := recover()
			status := writer.status
			if err == nil {
				l.Info("%d %s %s %s %s", status, method, url, remote, end.Sub(start))
			} else {
				l.Error("%d %s %s %s %s err:'%s'", status, method, url, remote, end.Sub(start),
					err)
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
