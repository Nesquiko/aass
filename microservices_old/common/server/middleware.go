package server

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	validation_middleware "github.com/oapi-codegen/nethttp-middleware"
)

type OapiValidationOptions struct {
	Spec         *openapi3.T
	ErrorHandler func(w http.ResponseWriter, message string, statusCode int)
}

type MiddlewareFunc func(http.Handler) http.Handler

func Middleware(logger *httplog.Logger, opts OapiValidationOptions) []MiddlewareFunc {
	return []MiddlewareFunc{
		chi_middleware.Recoverer,
		cors.Handler(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			MaxAge:         300,
		}),
		chi_middleware.RealIP,
		validation_middleware.OapiRequestValidatorWithOptions(
			opts.Spec,
			&validation_middleware.Options{ErrorHandler: opts.ErrorHandler},
		),
		httplog.RequestLogger(logger),
		chi_middleware.AllowContentType(ApplicationJSON),
	}
}

func Heartbeat() func(http.Handler) http.Handler {
	return chi_middleware.Heartbeat("/monitoring/heartbeat")
}

func OptionsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*") // Or specific origins
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().
				Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
				// Add any other headers your frontend sends
			w.Header().
				Set("Access-Control-Max-Age", "86400")
				// Cache preflight response for 1 day
			w.WriteHeader(
				http.StatusNoContent,
			) // 204 No Content is standard for OPTIONS
			return
		}
		// If not an OPTIONS request, pass it down the middleware chain
		next.ServeHTTP(w, r)
	})
}
