package main

import (
	"context"
	"fmt"
	"os"
	_ "time/tzdata"

	patientservice "github.com/Nesquiko/aass/patient-service"
)

func main() {
	ctx := context.Background()
	if err := patientservice.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
