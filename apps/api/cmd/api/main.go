package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/config"
	"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/database"
	"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/httpserver"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	health := httpserver.HealthHandler{DB: db}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health.Liveness)
	mux.HandleFunc("GET /readyz", health.Readiness)

	server := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout: 60 * time.Second,
	}

	log.Printf("sodienthoai api listening on :%s", cfg.Port)
	log.Fatal(server.ListenAndServe())
}
