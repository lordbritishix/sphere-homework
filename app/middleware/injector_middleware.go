package middleware

import (
	"context"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"sphere-homework/app/config"
	"sphere-homework/app/repository"
	"sphere-homework/app/services"
)

const LoggerKey = "logger"
const ConfigKey = "config"
const ServicesContextKey = "services"

// InjectorMiddleware injects logger and config into the request context
func InjectorMiddleware(
	logger *zap.Logger,
	config *config.Config,
	servicesContext *ServicesContext) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Attach to the request context
			ctx := context.WithValue(r.Context(), LoggerKey, logger)
			ctx = context.WithValue(r.Context(), ConfigKey, config)
			ctx = context.WithValue(r.Context(), ServicesContextKey, servicesContext)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetFeeRepository(r *http.Request) *repository.FeeRepository {
	s, ok := r.Context().Value(ServicesContextKey).(*ServicesContext)
	if !ok {
		return nil
	}
	return s.FeeRepository
}

func GetLedgerRepository(r *http.Request) *repository.LedgerRepository {
	s, ok := r.Context().Value(ServicesContextKey).(*ServicesContext)
	if !ok {
		return nil
	}
	return s.LedgerRepository
}

func GetRateRepository(r *http.Request) *repository.RateRepository {
	s, ok := r.Context().Value(ServicesContextKey).(*ServicesContext)
	if !ok {
		return nil
	}
	return s.RateRepository
}

func GetEventService(r *http.Request) *services.EventService {
	s, ok := r.Context().Value(ServicesContextKey).(*ServicesContext)
	if !ok {
		return nil
	}
	return s.EventService
}

func GetLogger(r *http.Request) *zap.Logger {
	logger, ok := r.Context().Value(LoggerKey).(*zap.Logger)
	if !ok {
		logger, err := zap.NewProduction()
		if err != nil {
			return nil
		}
		return logger
	}
	return logger
}

func GetConfig(r *http.Request) *config.Config {
	c, ok := r.Context().Value(ConfigKey).(*config.Config)
	if !ok {
		cfg := config.NewConfig()
		return &cfg
	}
	return c
}
