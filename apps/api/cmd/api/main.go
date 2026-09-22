package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/cache"
	"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/config"
	"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/database"
	"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/httpserver"
	"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/phone"
	"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/search"
)

func main() {
	cfg, err := config.Load()
	if err != nil { log.Fatal(err) }

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil { log.Fatalf("database connection failed: %v", err) }
	defer db.Close()

	redisClient, err := cache.Open(ctx, cfg.RedisAddr)
	if err != nil { log.Fatalf("redis connection failed: %v", err) }
	defer redisClient.Close()

	searchClient, err := search.Open(ctx, cfg.OpenSearchURL)
	if err != nil { log.Fatalf("opensearch connection failed: %v", err) }

	health := httpserver.HealthHandler{DB: db, Redis: redisClient, Search: searchClient}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health.Liveness)
	mux.HandleFunc("GET /readyz", health.Readiness)
	phoneHandler := httpserver.PhoneHandler{Repository: phone.Repository{DB: db}}
	mux.HandleFunc("GET /v1/phone/{number}", phoneHandler.Get)

	server := &http.Server{
		Addr: ":" + cfg.Port, Handler: mux,
		ReadHeaderTimeout: 5*time.Second, ReadTimeout: 10*time.Second,
		WriteTimeout: 15*time.Second, IdleTimeout: 60*time.Second,
	}
	log.Printf("sodienthoai api listening on :%s", cfg.Port)
	log.Fatal(server.ListenAndServe())
}
