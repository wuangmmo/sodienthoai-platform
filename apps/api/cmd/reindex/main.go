package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/search"
)

type document struct {
	ID string `json:"id"`
	E164 string `json:"e164"`
	CountryCode string `json:"country_code"`
	CallingCode string `json:"calling_code"`
	NationalNumber string `json:"national_number"`
	NumberType *string `json:"number_type,omitempty"`
	VerificationStatus string `json:"verification_status"`
	SEOStatus string `json:"seo_status"`
	SpamScore float64 `json:"spam_score"`
	ReportCount int64 `json:"report_count"`
	SearchCount int64 `json:"search_count"`
	DataQualityScore float64 `json:"data_quality_score"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

func main() {
	ctx := context.Background()
	dbURL, searchURL := os.Getenv("DATABASE_URL"), os.Getenv("OPENSEARCH_URL")
	if dbURL=="" || searchURL=="" { log.Fatal("DATABASE_URL and OPENSEARCH_URL are required") }
	batchSize:=1000
	if raw:=os.Getenv("REINDEX_BATCH_SIZE"); raw!="" {
		n,err:=strconv.Atoi(raw); if err!=nil || n<1 || n>5000 { log.Fatal("REINDEX_BATCH_SIZE must be between 1 and 5000") }
		batchSize=n
	}

	db,err:=sql.Open("pgx",dbURL); if err!=nil { log.Fatal(err) }; defer db.Close()
	client,err:=search.Open(ctx,searchURL); if err!=nil { log.Fatal(err) }
	if err:=client.EnsurePhoneIndex(ctx); err!=nil { log.Fatal(err) }

	rows,err:=db.QueryContext(ctx,`SELECT id::text,e164,country_code,calling_code,national_number,number_type,
verification_status::text,seo_status::text,spam_score::float8,report_count,search_count,
data_quality_score::float8,last_seen_at FROM phone_numbers ORDER BY id`)
	if err!=nil { log.Fatal(err) }; defer rows.Close()

	var buf bytes.Buffer
	count,batch:=0,0
	flush:=func() {
		if batch==0 { return }
		if err:=client.BulkIndexPhones(ctx,buf.Bytes()); err!=nil { log.Fatal(err) }
		count+=batch; batch=0; buf.Reset()
	}

	for rows.Next() {
		var d document
		if err:=rows.Scan(&d.ID,&d.E164,&d.CountryCode,&d.CallingCode,&d.NationalNumber,&d.NumberType,
			&d.VerificationStatus,&d.SEOStatus,&d.SpamScore,&d.ReportCount,&d.SearchCount,&d.DataQualityScore,&d.LastSeenAt); err!=nil { log.Fatal(err) }
		meta,_:=json.Marshal(map[string]any{"index":map[string]string{"_index":search.PhoneIndex,"_id":d.ID}})
		body,err:=json.Marshal(d); if err!=nil { log.Fatal(err) }
		buf.Write(meta); buf.WriteByte('\n'); buf.Write(body); buf.WriteByte('\n')
		batch++
		if batch>=batchSize { flush() }
	}
	if err:=rows.Err(); err!=nil { log.Fatal(err) }
	flush()
	log.Printf("reindexed %d phone numbers in batches of up to %d",count,batchSize)
}
