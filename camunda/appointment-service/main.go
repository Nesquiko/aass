package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	_ "time/tzdata"

	"github.com/Nesquiko/aass/appointment-service/api"
	"github.com/Nesquiko/aass/common/server"
)

func main() {
	ctx := context.Background()

	spec, err := api.GetSwagger()
	if err != nil {
		slog.Error("failed to load OpenApi spec", slog.String("error", err.Error()))
		os.Exit(1)
	}

	var dbProvider server.MongoDbProvider[mongoAppointmentDb] = newMongoAppointmentDb
	var serverProvider server.ServerProvider[mongoAppointmentDb] = newAppointmentServer

	if err := server.Run(ctx, "appointment-service", "APPOINTMENTSERVICE", spec, serverProvider, dbProvider); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
