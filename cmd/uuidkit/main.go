package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/relentlessworks/uuidkit/internal/api"
	"github.com/relentlessworks/uuidkit/internal/auth"
	"github.com/relentlessworks/uuidkit/internal/config"
	"github.com/relentlessworks/uuidkit/internal/store"
)

func main() {
	cfg := config.Load()

	s := store.New()

	h := api.New(s)

	mux := http.NewServeMux()

	h.RegisterRoutes(mux)

	var handler http.Handler = mux
	if cfg.APIKey != "" {
		handler = auth.APIKeyMiddleware(cfg.APIKey)(handler)
	}

	log.Printf("uuidkit listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, handler); err != nil {
		fmt.Fprintf(log.Writer(), "error: %v\n", err)
	}
}
