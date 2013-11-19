package app

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/gorilla/context"
	"net/http"
)

// csrf tokens
const CSRF_TOKEN_LEN = 64
const CONTEXT_CSRF_TOKEN = "csrf_token"
const FORM_CSRF_TOKEN = "CsrfToken"

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
