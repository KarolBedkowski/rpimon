package app

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
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
		csrfToken := session.Values[CONTEXTCSRFTOKEN]
		if r.Method == "POST" && r.FormValue(FORMCSRFTOKEN) != csrfToken {
			http.Error(w, "Fobidden/CSRF", http.StatusForbidden)
			//h.ServeHTTP(w, r)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

func createNewCsrfToken() string {
	token := make([]byte, CSRFTOKENLEN)
	rand.Read(token)
	csrfToken := base64.StdEncoding.EncodeToString(token)
	return csrfToken
}
