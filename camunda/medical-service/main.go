package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	_ "time/tzdata"

	"github.com/Nesquiko/aass/common/server"
	"github.com/Nesquiko/aass/medical-service/api"
)

func main() {
	ctx := context.Background()

	spec, err := api.GetSwagger()
	if err != nil {
		slog.Error("failed to load OpenApi spec", slog.String("error", err.Error()))
		os.Exit(1)
	}

	var dbProvider server.MongoDbProvider[mongoMedicalDb] = newMongoMedicalDb
	var serverProvider server.ServerProvider[mongoMedicalDb] = newMedicalServer

	if err := server.Run(ctx, "medical-service", "MEDICALSERVICE", spec, serverProvider, dbProvider); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
