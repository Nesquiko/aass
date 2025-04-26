package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"

	"github.com/Nesquiko/aass/common/server/api"
)

type Disconnecter interface {
	Disconnect(ctx context.Context) error
}

type (
	MongoDbProvider[DB Disconnecter] = func(ctx context.Context, uri string, db string) (DB, error)
	ServerProvider[DB Disconnecter]  = func(db DB, logger *httplog.Logger, opts api.ChiServerOptions) http.Handler
)

type ApiError struct {
	api.ErrorDetail
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("error %q, status %d", e.Title, e.Status)
}

func Run[DB Disconnecter](
	ctx context.Context,
	serviceName string,
	serviceEnvPrefix string,
	apiSpec *openapi3.T,
	serverProvider ServerProvider[DB],
	dbProvider MongoDbProvider[DB],
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cfg, err := LoadConfig(serviceEnvPrefix)
	if err != nil {
		slog.Error("failed to read config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	httpLogger := SetupLogger(serviceName, cfg.Log.Level)

	loc, err := time.LoadLocation(cfg.App.Timezone)
	if err != nil {
		slog.Error("failed to load timezone", slog.String("error", err.Error()))
		os.Exit(1)
	}
	httpLogger.Info("loaded timezone", slog.String("tz", loc.String()))
	time.Local = loc

	db, err := dbProvider(ctx, cfg.MongoURI(), cfg.Mongo.Db)
	if err != nil {
		slog.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	srv := NewServer(apiSpec, db, httpLogger, serverProvider)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(cfg.App.Host, cfg.App.Port),
		Handler: srv,
	}

	go func() {
		slog.Info("starting server", slog.String("addr", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("error listening and serving", slog.String("error", err.Error()))
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		slog.Info("interrupt received, shutting down server")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("error shutting down http server", slog.String("error", err.Error()))
		}
	}()
	wg.Wait()
	return nil
}

func SetupLogger(name string, logLevel slog.Level) *httplog.Logger {
	logger := httplog.NewLogger(name, httplog.Options{
		LogLevel: slog.Level(logLevel),
	})
	slog.SetDefault(logger.Logger)
	return logger
}

func NewServer[DB Disconnecter](
	spec *openapi3.T,
	db DB,
	middlewareLogger *httplog.Logger,
	serverProvider ServerProvider[DB],
) http.Handler {
	r := chi.NewMux()
	r.Use(Heartbeat())
	r.Use(OptionsMiddleware)

	validationOpts := OapiValidationOptions{
		Spec:         spec,
		ErrorHandler: validationErrorHandler,
	}

	serverMiddlewares := Middleware(middlewareLogger, validationOpts)
	apiMiddlewares := make([]api.MiddlewareFunc, len(serverMiddlewares))
	for i, mw := range serverMiddlewares {
		apiMiddlewares[i] = api.MiddlewareFunc(mw)
	}

	opts := api.ChiServerOptions{
		BaseRouter:  r,
		Middlewares: apiMiddlewares,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			var invalidParamErr *api.InvalidParamFormatError
			var requiredParamError *api.RequiredParamError

			switch {
			case errors.As(err, &invalidParamErr):
				slog.Warn(
					"invalid path param",
					slog.String("error", err.Error()),
					slog.String("where", "ErrorHandlerFunc"),
				)
				EncodeError(w, fromInvalidParamErr(invalidParamErr, chi.URLParam(r, "id")))
			case errors.As(err, &requiredParamError):
				slog.Warn(
					"missing required path param",
					slog.String("error", err.Error()),
					slog.String("where", "ErrorHandlerFunc"),
				)
				EncodeError(w, fromRequiredParamErr(requiredParamError))
			default:
				slog.Error(
					"unexpected error handling in ErrorHandlerFunc",
					slog.String("error", err.Error()),
				)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	}

	return serverProvider(db, middlewareLogger, opts)
}

func validationErrorHandler(w http.ResponseWriter, message string, statusCode int) {
	slog.Warn("validationErrorHandler", "message", message)

	if message == "no matching operation was found" {
		validationErr := &ApiError{
			ErrorDetail: api.ErrorDetail{
				Code:   NotFoundCode,
				Title:  "Route not found",
				Detail: "Route not found",
				Status: http.StatusNotFound,
			},
		}
		EncodeError(w, validationErr)
		return
	}

	split := strings.Split(message, ": ")
	hashIndex := strings.Index(split[1], "#")

	var schema *string = nil
	if hashIndex != -1 {
		s := split[1][hashIndex:]
		schema = &s
	}

	var path string
	if strings.HasPrefix(split[0], "parameter") {
		path = extractParamPath(split[0])
	} else {
		path = extractSchemaPath(split[2 : len(split)-1])
	}
	reason := split[len(split)-1]

	EncodeError(w, validationError(path, reason, statusCode, schema))
}

func extractSchemaPath(subPaths []string) string {
	path := ""
	for _, subPath := range subPaths {
		subPath = strings.TrimPrefix(subPath, "Error at")
		subPath = strings.TrimSpace(subPath)
		subPath = strings.Trim(subPath, "\"")
		path += subPath
	}
	return path
}

func extractParamPath(path string) string {
	p := strings.TrimPrefix(path, "parameter ")
	p = strings.TrimSuffix(p, "in query has an error")
	p = strings.TrimSpace(p)

	return p
}
