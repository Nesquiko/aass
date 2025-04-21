package pkg

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/Nesquiko/aass/common/server"
	"github.com/Nesquiko/aass/doctor-service/api"
	apptapi "github.com/Nesquiko/aass/doctor-service/api/appt-api"
)

func Run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cfg, err := LoadConfig()
	if err != nil {
		slog.Error("failed to read config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	httpLogger := server.SetupLogger("doctor-service", cfg.Log.Level)

	loc, err := time.LoadLocation(cfg.App.Timezone)
	if err != nil {
		slog.Error("failed to load timezone", slog.String("error", err.Error()))
		os.Exit(1)
	}
	httpLogger.Info("loaded timezone", slog.String("tz", loc.String()))
	time.Local = loc

	db, err := NewMongoDoctorDb(ctx, cfg.MongoURI(), cfg.Mongo.Db)
	if err != nil {
		slog.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	spec, err := api.GetSwagger()
	if err != nil {
		slog.Error("failed to load OpenApi spec", slog.String("error", err.Error()))
		os.Exit(1)
	}

	apptClient, _ := apptapi.NewClientWithResponses("http://doctor-service:8080/")
	app := DoctorApp{db: db, apptClient: apptClient}
	srv := NewServer(app, spec, httpLogger)

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
