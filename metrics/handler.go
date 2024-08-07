package metrics

import (
	"net/http"
	"strings"
	"time"
)

// Tailored Prometheus metrics
const (
	URLPrefix = "archivist"
)

// we have to intercept the ResponseWriter in order to get the statuscode
type LoggingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (m *Metrics) NewLatencyMetricsHandler(h http.Handler) http.Handler {

	if m == nil {
		return h
	}
	m.log.Debugf("NewLatencyMetricsHandler")
	observer := NewLatencyObservers(m)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := m.log.FromContext(r.Context())
		defer log.Close()

		log.Debugf("Request URL Path: '%s'", r.URL.Path)
		fields := strings.Split(strings.Trim(r.URL.Path, "/ "), "/")
		log.Debugf("Fields: %v (%d)", fields, len(fields))
		if fields[0] != URLPrefix {
			h.ServeHTTP(w, r)
			return
		}
		// generate pre-process metrics here...
		// nothing at the moment...

		// WriteHeader(int) is not called if our response implicitly returns 200 OK, so
		// we default to that status code.
		lrw := &LoggingResponseWriter{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		start := time.Now()
		// process original handler
		h.ServeHTTP(lrw, r)

		latency := time.Since(start).Seconds()

		// generate post-process metrics here...

		observer.ObserveRequestsCount(fields, r.Method)
		observer.ObserveRequestsLatency(latency, fields, r.Method)
	})
}
