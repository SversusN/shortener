package logger

import (
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type ServerLogger struct {
	logger *zap.Logger //не хочу делать Sugar, но хочу использовать ниже.
}

func CreateLogger(level zap.AtomicLevel) *ServerLogger {
	cfg := zap.NewProductionConfig()
	cfg.Level = level
	l, err := cfg.Build()

	if err != nil {
		return nil
	}
	return &ServerLogger{
		logger: l,
	}
}

type responseData struct {
	size   int
	status int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func (l ServerLogger) LoggingMW() func(http.Handler) http.Handler {
	sl := *l.logger.Sugar()
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			log.Print("Error syncing logger", zap.Error(err))
		}
	}(l.logger)
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, req *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					return
				}
			}()
			responseData := &responseData{
				size:   0,
				status: 0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}
			start := time.Now()
			next.ServeHTTP(&lw, req)
			duration := time.Since(start)
			sl.Infoln(
				"uri", req.RequestURI,
				"method", req.Method,
				"status", responseData.status,
				"duration", duration,
				"size", responseData.size,
			)
		}
		return http.HandlerFunc(fn)
	}
}
