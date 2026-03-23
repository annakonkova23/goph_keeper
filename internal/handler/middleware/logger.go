// package middleware - для middleware http handler
package middleware

import (
	"bytes"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"time"
)

type responseLogger struct {
	http.ResponseWriter
	status int
	size   int
}

func (l *responseLogger) WriteHeader(code int) {
	if l.status != 0 {
		return
	}
	l.status = code
	l.ResponseWriter.WriteHeader(code)
}

func (l *responseLogger) Write(b []byte) (int, error) {
	if l.status == 0 {
		l.status = http.StatusOK
	}
	size, err := l.ResponseWriter.Write(b)
	l.size += size
	return size, err
}

// LoggingMiddleware - логирование запроса и ответа.
func LoggingMiddleware(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			uri := r.URL.RequestURI()
			method := r.Method
			contentType := r.Header.Get("Content-Type")

			var bodyBytes []byte
			if r.Body != nil && (contentType == "" || !strings.HasPrefix(contentType, "multipart/")) {
				bodyBytes, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
			log := log.With(zap.String("method", method),
				zap.String("uri", uri))
			log.Info("Получен запрос",
				zap.String("content_type", contentType),
				zap.ByteString("body", bodyBytes))

			logger := &responseLogger{ResponseWriter: w}

			next.ServeHTTP(logger, r)

			log.Info("Запрос выполнен",
				zap.Int("status", logger.status),
				zap.Duration("time", time.Since(start)),
				zap.Int("size", logger.size),
			)
		})
	}
}
