//go:build !solution

package requestlog

import (
	"github.com/felixge/httpsnoop"
	"net/http"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

var id = atomic.Int64{}

func Log(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			curID := id.Load()
			id.Add(1)
			l.Info("request started", zap.String("path", r.URL.Path), zap.String("method", r.Method), zap.Int64("request_id", curID))
			statusCode := http.StatusOK
			wrappedW := httpsnoop.Wrap(w, httpsnoop.Hooks{WriteHeader: func(h httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
				return func(code int) {
					statusCode = code
					h(code)
				}
			}})
			startTime := time.Now()
			defer func() {
				endTime := time.Now()
				re := recover()
				if re != nil {
					l.Info("request panicked", zap.String("path", r.URL.Path),
						zap.String("method", r.Method), zap.Int64("request_id", curID),
						zap.Duration("duration", endTime.Sub(startTime)),
						zap.Int("status_code", statusCode))
					panic(re.(error))
				}
			}()
			next.ServeHTTP(wrappedW, r)
			endTime := time.Now()
			l.Info("request finished", zap.String("path", r.URL.Path),
				zap.String("method", r.Method), zap.Int64("request_id", curID),
				zap.Duration("duration", endTime.Sub(startTime)),
				zap.Int("status_code", statusCode))
		})
	}
}
