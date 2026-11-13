package main

import (
	"context"
	"fmt"
	"os"

	"server-monitor-api/internal/services"
)

func main() {
	// Retrieve services
	servicesJSON, err := services.ListServices(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "list services: %v\n", err)
		os.Exit(1)
	}

	// Output services
	if _, writeErr := fmt.Fprintln(os.Stdout, servicesJSON); writeErr != nil {
		fmt.Fprintf(os.Stderr, "write services: %v\n", writeErr)
		os.Exit(1)
	}
}
