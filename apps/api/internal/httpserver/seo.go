package httpserver

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/phone"
)

type SEOHandler struct{ Repository phone.Repository }

func (h SEOHandler) Sitemap(w http.ResponseWriter,r *http.Request){
	limit:=50000
	if raw:=r.URL.Query().Get("limit");raw!="" { if n,err:=strconv.Atoi(raw);err==nil { limit=n } }
	ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel()
	items,err:=h.Repository.Sitemap(ctx,limit)
	if err!=nil { writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"});return }
	writeJSON(w,http.StatusOK,map[string]any{"data":items})
}
