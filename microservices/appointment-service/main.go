package main

import (
	"context"
	"fmt"
	"os"
	_ "time/tzdata"

	"github.com/Nesquiko/aass/appointment-service/pkg"
)

func main() {
	ctx := context.Background()
	if err := pkg.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
