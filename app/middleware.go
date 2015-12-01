package app

import (
	"fmt"
	l "k.prv/rpimon/logging"
	"net/http"
	"runtime/debug"
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

// logHandler log all requests.
func logHandler(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		writer := &loggingResponseWriter{ResponseWriter: w, status: 200}
		defer func() {
			end := time.Now()
			stack := debug.Stack()
			if err := recover(); err == nil {
				l.Debug("%d %s %s %s %s", writer.status, r.Method, r.URL.String(), r.RemoteAddr, end.Sub(start))
			} else {
				l.Error(fmt.Sprint("%d %s %s %s %s err:'%#v'\n%s",
					writer.status, r.Method, r.URL.String(), r.RemoteAddr,
					end.Sub(start), err, stack))
			}
		}()
		h.ServeHTTP(writer, r)
	})
}

func TimeoutHandler(h http.HandlerFunc, sec int) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		done := make(chan bool, 1)
		go func() {
			h(w, r)
			done <- true
		}()
		select {
		case <-done:
			return
		case <-time.After(time.Duration(sec) * time.Second):
			http.Error(w, "Timeout", http.StatusInternalServerError)
			return
		}
	})
}
