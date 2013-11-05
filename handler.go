package loghandler

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	statNameHeader = "X-Stat"
)

// loggingHandler is the http.Handler implementation for LoggingHandlerTo and its friends
type loggingHandler struct {
	writer  io.Writer
	handler http.Handler
}

func (h loggingHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	t := time.Now()
	logger := responseLogger{w: w}
	h.handler.ServeHTTP(&logger, req)
	statName := logger.Header().Get(statNameHeader)
	writeLog(h.writer, req, t, logger.status, logger.size, statName)
}

// responseLogger is wrapper of http.ResponseWriter that keeps track of its HTTP status
// code and body size
type responseLogger struct {
	w        http.ResponseWriter
	status   int
	size     int
	statName string
}

func (l *responseLogger) Header() http.Header {
	return l.w.Header()
}

func (l *responseLogger) Write(b []byte) (int, error) {
	if l.status == 0 {
		// The status will be StatusOK if WriteHeader has not been called yet
		l.status = http.StatusOK
	}
	size, err := l.w.Write(b)
	l.size += size
	return size, err
}

func (l *responseLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

// buildCommonLogLine builds a log entry for req in Apache Common Log Format.
// ts is the timestamp with which the entry should be logged.
// status and size are used to provide the response HTTP status and size.
func buildCommonLogLine(req *http.Request, ts time.Time, status int, size int, statName string) string {
	username := "-"
	if req.URL.User != nil {
		if name := req.URL.User.Username(); name != "" {
			username = name
		}
	}

	return fmt.Sprintf("%s - %s [%s] \"%s %s %s\" %d %d (%s)",
		strings.Split(req.RemoteAddr, ":")[0],
		username,
		ts.Format("02/Jan/2006:15:04:05 -0700"),
		req.Method,
		req.URL.RequestURI(),
		req.Proto,
		status,
		size,
		statName,
	)
}

// writeLog writes a log entry for req to w in Apache Common Log Format.
// ts is the timestamp with which the entry should be logged.
// status and size are used to provide the response HTTP status and size.
func writeLog(w io.Writer, req *http.Request, ts time.Time, status, size int, statName string) {
	line := buildCommonLogLine(req, ts, status, size, statName) + "\n"
	fmt.Fprint(w, line)
}

// LoggingHandler return a http.Handler that wraps h and logs requests to out in
// Apache Common Log Format (CLF).
//
// See http://httpd.apache.org/docs/2.2/logs.html#common for a description of this format.
//
// LoggingHandler always sets the ident field of the log to -
func NewLoggingHandler(out io.Writer, h http.Handler) http.Handler {
	return loggingHandler{out, h}
}

func SetStat(w http.ResponseWriter, name string) {
	w.Header().Add(statNameHeader, name)
}
