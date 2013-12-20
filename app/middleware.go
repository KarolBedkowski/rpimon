package app

import (
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"runtime/debug"
	"time"
)

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
