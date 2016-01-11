package app

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
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

	httpResponseTime := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_response_time_miliseconds",
			Help: "The HTTP request latencies in miliseconds.",
		},
		[]string{"url", "handler"},
	)

	prometheus.MustRegister(httpResponseTime)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		writer := &loggingResponseWriter{ResponseWriter: w, status: 200}
		defer func() {
			end := time.Now()
			stack := debug.Stack()
			duration := end.Sub(start)
			if err := recover(); err == nil {
				l.Debug("%d %s %s %s %s", writer.status, r.Method, r.URL.String(), r.RemoteAddr, duration)
			} else {
				l.Error(fmt.Sprint("%d %s %s %s %s err:'%#v'\n%s",
					writer.status, r.Method, r.URL.String(), r.RemoteAddr,
					duration, err, stack))
			}
			httpResponseTime.WithLabelValues(r.URL.Path, "rpimon").Observe(duration.Seconds() * 1000)
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
