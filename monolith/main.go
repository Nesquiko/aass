package main

import (
	"context"
	"fmt"
	"os"
	_ "time/tzdata"

	"github.com/Nesquiko/wac/pkg/server"
)

func main() {
	ctx := context.Background()
	if err := server.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
