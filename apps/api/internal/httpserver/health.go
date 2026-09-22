package httpserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/search"
)

type HealthHandler struct {
	DB     *sql.DB
	Redis  *redis.Client
	Search *search.Client
}

func (h HealthHandler) Liveness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status":"ok","service":"sodienthoai-api","version":"0.1.0"})
}

func (h HealthHandler) Readiness(w http.ResponseWriter, _ *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	status := map[string]string{"database":"ok","redis":"ok","search":"ok"}
	ready := true
	if h.DB == nil || h.DB.PingContext(ctx) != nil { status["database"]="unavailable"; ready=false }
	if h.Redis == nil || h.Redis.Ping(ctx).Err() != nil { status["redis"]="unavailable"; ready=false }
	if h.Search == nil || h.Search.Ping(ctx) != nil { status["search"]="unavailable"; ready=false }

	if !ready {
		status["status"]="not_ready"
		writeJSON(w,http.StatusServiceUnavailable,status)
		return
	}
	status["status"]="ready"
	writeJSON(w,http.StatusOK,status)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
