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
	limit:=50000; page:=0
	if raw:=r.URL.Query().Get("limit");raw!="" { if n,err:=strconv.Atoi(raw);err==nil { limit=n } }
	if raw:=r.URL.Query().Get("page");raw!="" { if n,err:=strconv.Atoi(raw);err==nil&&n>=0 { page=n } }
	if limit<1||limit>50000 { limit=50000 }
	ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel()
	items,err:=h.Repository.SitemapPage(ctx,limit,page*limit)
	if err!=nil { writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"});return }
	writeJSON(w,http.StatusOK,map[string]any{"data":items,"page":page,"limit":limit})
}

func (h SEOHandler) SitemapCount(w http.ResponseWriter,r *http.Request){
	ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel()
	count,err:=h.Repository.SitemapCount(ctx)
	if err!=nil { writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"});return }
	writeJSON(w,http.StatusOK,map[string]any{"count":count,"page_size":50000})
}
