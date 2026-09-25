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
	}, ReporterHashSecret: cfg.ReporterHashSecret}
	phoneLookup := httpserver.RateLimitByIP(phoneHandler.Get, 120, time.Minute)
	mux.HandleFunc("GET /v1/phone/{number}", phoneLookup)
	mux.HandleFunc("GET /v1/phone/{number}/profile", httpserver.RateLimitByIP(phoneHandler.Profile, 120, time.Minute))
	mux.HandleFunc("GET /v1/phone/{number}/footprint", httpserver.RateLimitByIP(phoneHandler.Footprint, 60, time.Minute))
	mux.HandleFunc("POST /v1/phone/{number}/footprint/scan", httpserver.RateLimitByIP(phoneHandler.RequestFootprintScan, 5, time.Hour))
	searchHandler := httpserver.SearchHandler{Repository: phone.Repository{DB: db}, Search: searchClient}
	mux.HandleFunc("GET /v1/search", httpserver.RateLimitByIP(searchHandler.Get, 120, time.Minute))
	mux.HandleFunc("POST /v1/phone/{number}/reports", httpserver.RateLimitByIP(phoneHandler.Report, 20, time.Hour))
	mux.HandleFunc("POST /v1/phone/{number}/claims", httpserver.RateLimitByIP(phoneHandler.Claim, 10, time.Hour))
	moderation := httpserver.ModerationHandler{Service: phoneHandler.Service, Token: cfg.AdminAPIToken}
	mux.HandleFunc("PATCH /v1/admin/reports/{id}", moderation.Report)
	mux.HandleFunc("PATCH /v1/admin/claims/{id}", moderation.Claim)
	mux.HandleFunc("PATCH /v1/admin/comments/{id}", moderation.Comment)
	mux.HandleFunc("POST /v1/admin/claims/{id}/evidence", moderation.Evidence)
	mux.HandleFunc("PATCH /v1/admin/appeals/{id}", moderation.Appeal)
	mux.HandleFunc("PATCH /v1/admin/business-verifications/{id}", moderation.BusinessVerification)
	mux.HandleFunc("PATCH /v1/admin/business-reviews/{id}", moderation.BusinessReview)
	admin := httpserver.AdminHandler{DB: db, Token: cfg.AdminAPIToken}
	mux.HandleFunc("GET /v1/admin/dashboard", admin.Dashboard)
	mux.HandleFunc("GET /v1/admin/reports", admin.Reports)
	mux.HandleFunc("GET /v1/admin/comments", admin.Comments)
	mux.HandleFunc("GET /v1/admin/phones", admin.Phones)
	mux.HandleFunc("GET /v1/admin/operations", admin.Operations)
	mux.HandleFunc("GET /v1/admin/claims", admin.Claims)
	mux.HandleFunc("GET /v1/admin/audit", admin.Audit)
	mux.HandleFunc("GET /v1/admin/users", admin.Users)
	mux.HandleFunc("GET /v1/admin/footprint", admin.Footprint)
	mux.HandleFunc("GET /v1/admin/identities", admin.Identities)
	mux.HandleFunc("GET /v1/admin/appeals", admin.Appeals)
	mux.HandleFunc("GET /v1/admin/business-verifications", admin.BusinessVerifications)
	mux.HandleFunc("GET /v1/admin/business-reviews", admin.BusinessReviews)
	adminImport := httpserver.AdminImportHandler{Service: phoneHandler.Service, Token: cfg.AdminAPIToken}
	mux.HandleFunc("POST /v1/admin/import", adminImport.Import)
	mux.HandleFunc("GET /v1/admin/import/batches", adminImport.Batches)
	contacts := httpserver.ContactHandler{Service: phoneHandler.Service}
	mux.HandleFunc("GET /v1/contacts", contacts.List)
	mux.HandleFunc("POST /v1/contacts/import", httpserver.RateLimitByIP(contacts.Import, 10, time.Hour))
	mux.HandleFunc("POST /v1/contacts/import-file", httpserver.RateLimitByIP(contacts.ImportFile, 10, time.Hour))
	mux.HandleFunc("GET /v1/contacts/alerts", contacts.Alerts)
	mux.HandleFunc("POST /v1/contacts/rescan", httpserver.RateLimitByIP(contacts.Rescan, 10, time.Hour))
	mux.HandleFunc("POST /v1/contacts/{id}/actions", httpserver.RateLimitByIP(contacts.Action, 30, time.Hour))
	mux.HandleFunc("DELETE /v1/contacts", contacts.DeleteAll)
	community := httpserver.CommunityHandler{Service: phoneHandler.Service}
	mux.HandleFunc("POST /v1/phone/{number}/follow", httpserver.RateLimitByIP(community.Follow, 60, time.Hour))
	mux.HandleFunc("GET /v1/phone/{number}/comments", httpserver.RateLimitByIP(community.ListComments, 120, time.Minute))
	mux.HandleFunc("POST /v1/phone/{number}/comments", httpserver.RateLimitByIP(community.Comment, 20, time.Hour))
	mux.HandleFunc("POST /v1/comments/{id}/helpful", httpserver.RateLimitByIP(community.Helpful, 60, time.Hour))
	mux.HandleFunc("POST /v1/comments/{id}/report", httpserver.RateLimitByIP(community.ReportComment, 20, time.Hour))
	mux.HandleFunc("GET /v1/me/follows", community.Follows)
	mux.HandleFunc("GET /v1/me/notifications", community.Notifications)
	mux.HandleFunc("PATCH /v1/me/notifications/{id}/read", community.MarkNotificationRead)
	mux.HandleFunc("GET /v1/me/notifications/summary", community.NotificationSummary)
	mux.HandleFunc("PATCH /v1/me/notifications/read-all", community.MarkAllNotificationsRead)
	accountParity := httpserver.AccountParityHandler{Service: phoneHandler.Service}
	mux.HandleFunc("GET /v1/me/phone-preferences", accountParity.Preferences)
	mux.HandleFunc("PUT /v1/me/phone-preferences/{number}", httpserver.RateLimitByIP(accountParity.SetPreference, 60, time.Hour))
	mux.HandleFunc("DELETE /v1/me/phone-preferences/{number}", accountParity.DeletePreference)
	mux.HandleFunc("GET /v1/me/lookup-history", accountParity.History)
	mux.HandleFunc("DELETE /v1/me/lookup-history", accountParity.DeleteHistory)
	mux.HandleFunc("POST /v1/phone/{number}/appeals", httpserver.RateLimitByIP(accountParity.Appeal, 5, time.Hour))
	mux.HandleFunc("GET /v1/me/appeals", accountParity.Appeals)
	mux.HandleFunc("DELETE /v1/me/appeals/{id}", accountParity.WithdrawAppeal)
	business := httpserver.BusinessHandler{Service: phoneHandler.Service}
	mux.HandleFunc("POST /v1/businesses", httpserver.RateLimitByIP(business.Create, 10, time.Hour))
	mux.HandleFunc("GET /v1/me/businesses", business.Mine)
	mux.HandleFunc("GET /v1/me/business-verifications", business.MyVerifications)
	mux.HandleFunc("POST /v1/businesses/{id}/verification-requests", httpserver.RateLimitByIP(business.Verification, 5, time.Hour))
	mux.HandleFunc("POST /v1/businesses/{id}/reviews", httpserver.RateLimitByIP(business.Review, 10, time.Hour))
	publicBusiness := httpserver.PublicBusinessHandler{Service: phoneHandler.Service}
	mux.HandleFunc("GET /v1/businesses/{slug}", httpserver.RateLimitByIP(publicBusiness.Get, 120, time.Minute))
	seoHandler := httpserver.SEOHandler{Repository: phone.Repository{DB: db}}
	mux.HandleFunc("GET /v1/seo/sitemap", seoHandler.Sitemap)
	mux.HandleFunc("GET /v1/seo/sitemap/count", seoHandler.SitemapCount)
	mux.HandleFunc("GET /v1/seo/business-sitemap", seoHandler.BusinessSitemap)

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
