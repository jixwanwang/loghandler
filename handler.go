package loghandler

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	statNameHeader = "X-Stat-Key"
)

// loggingHandler is the http.Handler implementation for LoggingHandlerTo and its friends
type loggingHandler struct {
	writer  io.Writer
	sl      StatsLogger
	handler http.Handler
}

func (h loggingHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger := responseLogger{
		w:     w,
		start: time.Now(),
	}
	h.handler.ServeHTTP(&logger, req)
	writeLog(h.writer, req, logger.start, logger.status,
		logger.size, logger.duration)

	statName := logger.Header().Get(statNameHeader)
	if h.sl != nil {
		h.sl.Timing(fmt.Sprintf("%s", statName), logger.duration)
		h.sl.IncrBy(fmt.Sprintf("%s.%d", statName, logger.status), 1)
	}
}

// responseLogger is wrapper of http.ResponseWriter that keeps track of its HTTP status
// code and body size
type responseLogger struct {
	w        http.ResponseWriter
	status   int
	size     int
	start    time.Time
	duration time.Duration
	statName string
}

func (l *responseLogger) Header() http.Header {
	return l.w.Header()
}

// Write stores the status and duration of the request. We don't track
// the time the request is being sent to the client as it's the client's
// responsibility to be fast.
func (l *responseLogger) Write(b []byte) (int, error) {
	if l.status == 0 {
		// The status will be StatusOK if WriteHeader has not been called yet
		l.status = http.StatusOK
	}

	l.duration = time.Now().Sub(l.start)

	size, err := l.w.Write(b)
	l.size += size
	return size, err
}

func (l *responseLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

// CloseNotify implements the CloseNotifier interface.
// If the underlying ResponseWriter implements CloserNotifier, simply call that,
// otherwise
func (l *responseLogger) CloseNotify() <-chan bool {
	if cn, ok := l.w.(http.CloseNotifier); ok {
		return cn.CloseNotify()
	}
	return nil
}

// buildCommonLogLine builds a log entry for req in Apache Common Log Format.
// ts is the timestamp with which the entry should be logged.
// status and size are used to provide the response HTTP status and size.
func buildCommonLogLine(req *http.Request, ts time.Time, status int, size int,
	duration time.Duration) string {
	username := "-"
	if req.URL.User != nil {
		if name := req.URL.User.Username(); name != "" {
			username = name
		}
	}

	return fmt.Sprintf("%s - %s [%s] \"%s %s %s\" %d %d (%dμs)",
		strings.Split(req.RemoteAddr, ":")[0],
		username,
		ts.Format("02/Jan/2006:15:04:05 -0700"),
		req.Method,
		req.URL.RequestURI(),
		req.Proto,
		status,
		size,
		duration/time.Microsecond,
	)
}

// writeLog writes a log entry for req to w in Apache Common Log Format.
// ts is the timestamp with which the entry should be logged.
// status and size are used to provide the response HTTP status and size.
func writeLog(w io.Writer, req *http.Request, ts time.Time, status,
	size int, duration time.Duration) {
	line := buildCommonLogLine(req, ts, status, size, duration) + "\n"
	fmt.Fprint(w, line)
}

// LoggingHandler return a http.Handler that wraps h and logs requests to out in
// Apache Common Log Format (CLF).
//
// See http://httpd.apache.org/docs/2.2/logs.html#common for a description of this format.
//
// LoggingHandler always sets the ident field of the log to -
func NewLoggingHandler(out io.Writer, sl StatsLogger, h http.Handler) http.Handler {
	return loggingHandler{out, sl, h}
}

func SetStat(w http.ResponseWriter, name string) {
	w.Header().Add(statNameHeader, name)
}

type StatsLogger interface {
	Timing(key string, t time.Duration)
	IncrBy(key string, delta int)
}
