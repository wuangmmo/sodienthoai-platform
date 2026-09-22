package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	phoneHandler := httpserver.PhoneHandler{Service: phone.Service{
		Repository: phone.Repository{DB: db},
		Cache: redisClient,
		TTL: 10 * time.Minute,
	}}
	phoneLookup := httpserver.RateLimitByIP(phoneHandler.Get, 120, time.Minute)
	mux.HandleFunc("GET /v1/phone/{number}", phoneLookup)
	searchHandler := httpserver.SearchHandler{Repository: phone.Repository{DB: db}}
	mux.HandleFunc("GET /v1/search", httpserver.RateLimitByIP(searchHandler.Get, 120, time.Minute))
	mux.HandleFunc("POST /v1/phone/{number}/reports", httpserver.RateLimitByIP(phoneHandler.Report, 20, time.Hour))
	mux.HandleFunc("POST /v1/phone/{number}/claims", httpserver.RateLimitByIP(phoneHandler.Claim, 10, time.Hour))
	moderation := httpserver.ModerationHandler{Service: phoneHandler.Service, Token: os.Getenv("ADMIN_API_TOKEN")}
	mux.HandleFunc("PATCH /v1/admin/reports/{id}", moderation.Report)
	mux.HandleFunc("PATCH /v1/admin/claims/{id}", moderation.Claim)
	seoHandler := httpserver.SEOHandler{Repository: phone.Repository{DB: db}}
	mux.HandleFunc("GET /v1/seo/sitemap", seoHandler.Sitemap)
	mux.HandleFunc("GET /v1/seo/sitemap/count", seoHandler.SitemapCount)

	server := &http.Server{
		Addr: ":" + cfg.Port, Handler: mux,
		ReadHeaderTimeout: 5*time.Second, ReadTimeout: 10*time.Second,
		WriteTimeout: 15*time.Second, IdleTimeout: 60*time.Second,
	}
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("sodienthoai api listening on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed { serverErr <- err }
	}()
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-serverErr: log.Fatal(err)
	case <-sigCtx.Done():
	}
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil { log.Printf("server shutdown failed: %v", err) }
}
