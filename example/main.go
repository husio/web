package main

import (
	"log"
	"net/http"

	"golang.org/x/net/context"
)

func main() {
	ctx := context.Background()

	// attach to context all external components that should be available in
	// handlers
	ctx = WithDB(ctx)

	app := NewApp(ctx)

	if err := http.ListenAndServe("localhost:8000", app); err != nil {
		log.Fatalf("HTTP server error: %s", err)
	}
}
