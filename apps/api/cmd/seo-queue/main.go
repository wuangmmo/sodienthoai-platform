package main

import (
 "context"
 "log"
 "os"

 "github.com/wuangmmo/sodienthoai-platform/apps/api/internal/database"
)

func main() {
 ctx := context.Background()
 u := os.Getenv("DATABASE_URL")
 if u == "" { log.Fatal("DATABASE_URL required") }
 db, err := database.Open(ctx, u)
 if err != nil { log.Fatal(err) }
 defer db.Close()

 // Generic phone pages are discovered through XML sitemaps and normal crawling.
 // Do not mark queue entries done unless a real supported indexing transport exists.
 var pending int64
 if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM seo_index_queue WHERE status IN ('pending','failed') AND available_at<=NOW()`).Scan(&pending); err != nil {
  log.Fatal(err)
 }
 log.Printf("seo queue has %d pending records; sitemap discovery is authoritative, no generic external indexing API configured", pending)
}
