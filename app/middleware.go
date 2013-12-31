package app

import (
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"time"
)

// loggingResponseWriter response writer with status
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader store status of request
func (writer *loggingResponseWriter) WriteHeader(status int) {
	writer.ResponseWriter.WriteHeader(status)
	writer.status = status
}

// Logging middleware
func logHandler(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		writer := &loggingResponseWriter{ResponseWriter: w, status: 200}
		defer func() {
			end := time.Now()
			if err := recover(); err == nil {
				l.Debug("%d %s %s %s %s", writer.status, r.Method, r.URL.String(), r.RemoteAddr, end.Sub(start))
			} else {
				l.Error("%d %s %s %s %s err:'%#v'", writer.status, r.Method, r.URL.String(), r.RemoteAddr, end.Sub(start),
					err)
			}
		}()
		h.ServeHTTP(writer, r)
	})
}
