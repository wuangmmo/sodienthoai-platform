package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main(){
	url:=os.Getenv("DATABASE_URL"); if url=="" { log.Fatal("DATABASE_URL is required") }
	db,err:=sql.Open("pgx",url); if err!=nil { log.Fatal(err) }; defer db.Close()
	ctx,cancel:=context.WithTimeout(context.Background(),2*time.Minute); defer cancel()

	result,err:=db.ExecContext(ctx,`
UPDATE phone_numbers
SET
  seo_status = CASE
    WHEN verification_status = 'verified' AND data_quality_score >= 70 THEN 'indexable'::seo_status
    WHEN report_count >= 3 AND data_quality_score >= 60 THEN 'indexable'::seo_status
    ELSE 'noindex'::seo_status
  END,
  seo_reason = CASE
    WHEN verification_status = 'verified' AND data_quality_score >= 70 THEN 'verified_quality'
    WHEN report_count >= 3 AND data_quality_score >= 60 THEN 'community_signal'
    ELSE 'insufficient_quality'
  END,
  seo_evaluated_at = NOW(),
  updated_at = NOW()
WHERE seo_status <> 'indexed'
  AND (
    seo_evaluated_at IS NULL
    OR seo_evaluated_at < updated_at
  )
`)
	if err!=nil { log.Fatal(err) }
	n,_:=result.RowsAffected()
	log.Printf("evaluated SEO quality for %d phone numbers",n)
}
