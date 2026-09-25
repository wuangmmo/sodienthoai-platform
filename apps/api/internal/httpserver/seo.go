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
	limit:=int(phone.SitemapPageSize)
	if raw:=r.URL.Query().Get("limit");raw!="" { if n,err:=strconv.Atoi(raw);err==nil { limit=n } }
	if limit<1||limit>int(phone.SitemapPageSize) { limit=int(phone.SitemapPageSize) }
	if raw:=r.URL.Query().Get("shard");raw!="" { shard,err1:=strconv.Atoi(raw);shards,err2:=strconv.Atoi(r.URL.Query().Get("shards"));if err1!=nil||err2!=nil||shards!=256||shard<0||shard>=shards { writeJSON(w,http.StatusBadRequest,map[string]string{"error":"invalid_shard"});return };ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel();items,err:=h.Repository.SitemapShard(ctx,shard,shards,limit);if err!=nil { writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"});return };writeJSON(w,http.StatusOK,map[string]any{"data":items,"limit":limit,"shard":shard,"shards":shards});return }
	after:=r.URL.Query().Get("after")
	ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel()
	items,next,err:=h.Repository.SitemapPage(ctx,limit,after)
	if err!=nil { writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"});return }
	writeJSON(w,http.StatusOK,map[string]any{"data":items,"limit":limit,"next_cursor":next})
}

func (h SEOHandler) SitemapCount(w http.ResponseWriter,r *http.Request){
	ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel()
	count,err:=h.Repository.SitemapCount(ctx)
	if err!=nil { writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"});return }
	writeJSON(w,http.StatusOK,map[string]any{"count":count,"page_size":phone.SitemapPageSize})
}


func (h SEOHandler) BusinessSitemap(w http.ResponseWriter,r *http.Request){
 limit:=50000;if raw:=r.URL.Query().Get("limit");raw!=""{if n,err:=strconv.Atoi(raw);err==nil&&n>0&&n<=50000{limit=n}}
 ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel()
 rows,err:=h.Repository.DB.QueryContext(ctx,`SELECT slug,updated_at FROM businesses WHERE verification_status='verified' ORDER BY updated_at DESC LIMIT $1`,limit)
 if err!=nil{writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"});return};defer rows.Close()
 items:=[]map[string]any{};for rows.Next(){var slug string;var updated time.Time;if err:=rows.Scan(&slug,&updated);err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};items=append(items,map[string]any{"slug":slug,"updated_at":updated})}
 writeJSON(w,http.StatusOK,map[string]any{"data":items,"limit":limit})
}
