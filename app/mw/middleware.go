package mw

import (
	"github.com/gorilla/context"
	"k.prv/rpimon/app/session"
	l "k.prv/rpimon/helpers/logging"
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

// Logging middleware
func LogHandler(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		writer := &loggingResponseWriter{ResponseWriter: w, status: 200}
		defer func() {
			end := time.Now()
			stack := debug.Stack()
			if err := recover(); err == nil {
				l.Debug("%d %s %s %s %s", writer.status, r.Method, r.URL.String(), r.RemoteAddr, end.Sub(start))
			} else {
				l.Error("%d %s %s %s %s err:'%#v'\n%s", writer.status, r.Method, r.URL.String(), r.RemoteAddr, end.Sub(start),
					err, stack)
			}
		}()
		h.ServeHTTP(writer, r)
	})
}

func SessionHandler(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := session.GetSessionStore(w, r)
		context.Set(r, "session", s)
		if ts, ok := s.Values[session.SessionTimestampKey]; ok {
			timestamp := time.Unix(ts.(int64), 0)
			now := time.Now()
			if now.Sub(timestamp) < session.MaxSessionAge {
				s.Values[session.SessionTimestampKey] = now.Unix()
			} else {
				s.Values = nil
			}
			s.Save(r, w)
		}
		//l.Debug("Context: %v", context.GetAll(r))
		h.ServeHTTP(w, r)
	})
}
