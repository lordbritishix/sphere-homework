package middleware

import (
	"bytes"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

func LoggerMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := GetLogger(r)
			var bodyBytes []byte
			if r.Body != nil {
				bodyBytes, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			start := time.Now()
			next.ServeHTTP(w, r)

			logger.Info("http request",
				zap.String("method", r.Method),
				zap.String("url", r.URL.String()),
				zap.Int64("duration ms", time.Since(start).Milliseconds()),
				zap.String("body", string(bodyBytes)),
			)

		})
	}
}
