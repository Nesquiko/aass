package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	_ "time/tzdata"

	"github.com/Nesquiko/aass/common/server"
	"github.com/Nesquiko/aass/user-service/api"
)

func main() {
	ctx := context.Background()

	spec, err := api.GetSwagger()
	if err != nil {
		slog.Error("failed to load OpenApi spec", slog.String("error", err.Error()))
		os.Exit(1)
	}

	var serverProvider server.ServerProvider[mongoUserDb] = newUserServer
	var dbProvider server.MongoDbProvider[mongoUserDb] = newMongoUserDb

	if err := server.Run(ctx, "user-service", "USERSERVICE", spec, serverProvider, dbProvider); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
