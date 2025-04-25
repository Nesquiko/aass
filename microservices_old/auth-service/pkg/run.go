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

	"github.com/Nesquiko/aass/auth-service/api"
	doctorapi "github.com/Nesquiko/aass/auth-service/api/doctor-api"
	patientapi "github.com/Nesquiko/aass/auth-service/api/patient-api"
	"github.com/Nesquiko/aass/common/server"
)

func Run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cfg, err := LoadConfig()
	if err != nil {
		slog.Error("failed to read config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	httpLogger := server.SetupLogger("auth-service", cfg.Log.Level)

	loc, err := time.LoadLocation(cfg.App.Timezone)
	if err != nil {
		slog.Error("failed to load timezone", slog.String("error", err.Error()))
		os.Exit(1)
	}
	httpLogger.Info("loaded timezone", slog.String("tz", loc.String()))
	time.Local = loc

	spec, err := api.GetSwagger()
	if err != nil {
		slog.Error("failed to load OpenApi spec", slog.String("error", err.Error()))
		os.Exit(1)
	}

	patientClient, _ := patientapi.NewClientWithResponses("http://patient-service:8080/")
	doctorClient, _ := doctorapi.NewClientWithResponses("http://doctor-service:8080/")
	app := AuthApp{patientClient: patientClient, doctorClient: doctorClient}
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
