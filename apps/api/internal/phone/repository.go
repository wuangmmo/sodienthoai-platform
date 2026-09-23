package phone

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrNotFound = errors.New("phone number not found")

type Repository struct{ DB *sql.DB }

func (r Repository) FindByE164(ctx context.Context, e164 string) (Number, error) {
	const q = `
SELECT id::text, country_code, calling_code, national_number, e164, number_type,
       verification_status::text, seo_status::text, spam_score::float8,
       report_count, search_count, data_quality_score::float8, first_seen_at, last_seen_at, reputation_label
FROM phone_numbers WHERE e164 = $1
`
	var n Number
	err := r.DB.QueryRowContext(ctx,q,e164).Scan(
		&n.ID,&n.CountryCode,&n.CallingCode,&n.NationalNumber,&n.E164,&n.NumberType,
		&n.VerificationStatus,&n.SEOStatus,&n.SpamScore,&n.ReportCount,&n.SearchCount,
		&n.DataQualityScore,&n.FirstSeenAt,&n.LastSeenAt,&n.ReputationLabel,
	)
	if errors.Is(err,sql.ErrNoRows) { return Number{},ErrNotFound }
	return n,err
}


type SitemapNumber struct {
	E164      string    `json:"e164"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r Repository) Sitemap(ctx context.Context, limit int) ([]SitemapNumber, error) {
	if limit < 1 || limit > 50000 { limit = 50000 }
	rows, err := r.DB.QueryContext(ctx, `
SELECT e164, updated_at
FROM phone_numbers
WHERE seo_status IN ('indexable','indexed')
ORDER BY updated_at DESC
LIMIT $1`, limit)
	if err != nil { return nil, err }
	defer rows.Close()
	items := make([]SitemapNumber,0)
	for rows.Next() {
		var item SitemapNumber
		if err := rows.Scan(&item.E164,&item.UpdatedAt); err != nil { return nil,err }
		items=append(items,item)
	}
	return items,rows.Err()
}


func (r Repository) SitemapCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM phone_numbers WHERE seo_status IN ('indexable','indexed')`).Scan(&count)
	return count, err
}

func (r Repository) SitemapPage(ctx context.Context, limit int, afterID string) ([]SitemapNumber, string, error) {
	if limit < 1 || limit > 50000 { limit = 50000 }
	rows, err := r.DB.QueryContext(ctx, `
SELECT id::text, e164, updated_at
FROM phone_numbers
WHERE seo_status IN ('indexable','indexed')
  AND ($2 = '' OR id > $2::uuid)
ORDER BY id
LIMIT $1`, limit, afterID)
	if err != nil { return nil, "", err }
	defer rows.Close()
	items := make([]SitemapNumber,0)
	next := ""
	for rows.Next(){var id string;var item SitemapNumber;if err:=rows.Scan(&id,&item.E164,&item.UpdatedAt);err!=nil{return nil,"",err};items=append(items,item);next=id}
	if err:=rows.Err();err!=nil{return nil,"",err}
	if len(items)<limit { next="" }
	return items,next,nil
}

func (r Repository) IdentitiesByPhoneID(ctx context.Context, phoneID string) ([]Identity, error) {
	rows, err := r.DB.QueryContext(ctx, `
SELECT id::text, kind::text, display_name, description, website_url, address_text,
       source_label, is_primary, confidence_score::float8
FROM phone_identities
WHERE phone_number_id = $1 AND is_public = TRUE
ORDER BY is_primary DESC, confidence_score DESC, created_at ASC`, phoneID)
	if err != nil { return nil, err }
	defer rows.Close()
	items := make([]Identity, 0)
	for rows.Next() {
		var i Identity
		if err := rows.Scan(&i.ID,&i.Kind,&i.DisplayName,&i.Description,&i.WebsiteURL,&i.AddressText,&i.SourceLabel,&i.IsPrimary,&i.ConfidenceScore); err != nil { return nil, err }
		items = append(items, i)
	}
	return items, rows.Err()
}

func (r Repository) RecordLookup(ctx context.Context, e164, countryCode string, phoneID *string, found bool) error {
	_, err := r.DB.ExecContext(ctx, `
INSERT INTO phone_lookup_events(phone_number_id,e164,country_code,found)
VALUES($1,$2,NULLIF($3,''),$4)`, phoneID, e164, countryCode, found)
	if err != nil { return err }
	if phoneID != nil {
		_, err = r.DB.ExecContext(ctx, `UPDATE phone_numbers SET search_count=search_count+1,last_seen_at=NOW() WHERE id=$1`, *phoneID)
	}
	return err
}

func (r Repository) EnsureDiscovered(ctx context.Context,e164,countryCode,callingCode,nationalNumber string)(Number,error){
 _,err:=r.DB.ExecContext(ctx,`
INSERT INTO phone_numbers(country_code,calling_code,national_number,e164,verification_status,seo_status)
VALUES($1,$2,$3,$4,'unverified','noindex')
ON CONFLICT(e164) DO UPDATE SET last_seen_at=NOW()`,countryCode,callingCode,nationalNumber,e164)
 if err!=nil{return Number{},err}
 return r.FindByE164(ctx,e164)
}

func (r Repository) RefreshDerivedStatus(ctx context.Context, phoneID string) error {
 n,err:=r.FindByID(ctx,phoneID);if err!=nil{return err}
 ids,err:=r.IdentitiesByPhoneID(ctx,phoneID);if err!=nil{return err}
 trust:=TrustScore(n,ids)
 _,err=r.DB.ExecContext(ctx,"UPDATE phone_numbers SET seo_status=$2,trust_score=$3,reputation_updated_at=NOW(),updated_at=NOW() WHERE id=$1",phoneID,SEOStatusFor(n,ids),trust)
 return err
}
func (r Repository) FindByID(ctx context.Context,id string)(Number,error){
 const q="SELECT id::text,country_code,calling_code,national_number,e164,number_type,verification_status::text,seo_status::text,spam_score::float8,report_count,search_count,data_quality_score::float8,first_seen_at,last_seen_at,reputation_label FROM phone_numbers WHERE id=$1"
 var n Number;err:=r.DB.QueryRowContext(ctx,q,id).Scan(&n.ID,&n.CountryCode,&n.CallingCode,&n.NationalNumber,&n.E164,&n.NumberType,&n.VerificationStatus,&n.SEOStatus,&n.SpamScore,&n.ReportCount,&n.SearchCount,&n.DataQualityScore,&n.FirstSeenAt,&n.LastSeenAt,&n.ReputationLabel)
 if errors.Is(err,sql.ErrNoRows){return Number{},ErrNotFound};return n,err
}


func (r Repository) ProfileSignalsByPhoneID(ctx context.Context, phoneID string) (ProfileSignals, error) {
	var s ProfileSignals
	err := r.DB.QueryRowContext(ctx, `
SELECT identity_confidence::float8, source_diversity, footprint_source_count,
       footprint_domain_count, footprint_category_count,
       footprint_first_detected_at, footprint_last_detected_at
FROM phone_profile_signals WHERE phone_number_id=$1`, phoneID).Scan(
		&s.IdentityConfidence,&s.SourceDiversity,&s.FootprintSourceCount,
		&s.FootprintDomainCount,&s.FootprintCategoryCount,
		&s.FirstDetectedAt,&s.LastDetectedAt)
	if errors.Is(err, sql.ErrNoRows) { return ProfileSignals{}, nil }
	return s, err
}

func (r Repository) HasOpenIdentityDispute(ctx context.Context, phoneID string) (bool, error) {
	var exists bool
	err := r.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM phone_identity_disputes WHERE phone_number_id=$1 AND status='open')`, phoneID).Scan(&exists)
	return exists, err
}


func (r Repository) FootprintByPhoneID(ctx context.Context, phoneID string, limit int) (FootprintSummary, error) {
 if limit<1||limit>100 { limit=25 }
 summary:=FootprintSummary{Categories:map[string]int64{},Occurrences:[]WebOccurrence{}}
 err:=r.DB.QueryRowContext(ctx,`SELECT COUNT(*),COUNT(DISTINCT domain),COUNT(DISTINCT category),MIN(detected_at),MAX(last_seen_at) FROM phone_web_occurrences WHERE phone_number_id=$1 AND is_active=TRUE`,phoneID).Scan(&summary.OccurrenceCount,&summary.IndependentDomainCount,&summary.CategoryCount,&summary.FirstDetectedAt,&summary.LastSeenAt)
 if err!=nil{return summary,err}
 rows,err:=r.DB.QueryContext(ctx,`SELECT category,COUNT(*) FROM phone_web_occurrences WHERE phone_number_id=$1 AND is_active=TRUE GROUP BY category ORDER BY COUNT(*) DESC`,phoneID);if err!=nil{return summary,err}
 for rows.Next(){var category string;var count int64;if err:=rows.Scan(&category,&count);err!=nil{rows.Close();return summary,err};summary.Categories[category]=count};if err:=rows.Close();err!=nil{return summary,err}
 items,err:=r.DB.QueryContext(ctx,`SELECT id::text,domain,url,page_title,category,context_snippet,source_confidence::float8,published_at,detected_at,last_seen_at,last_checked_at FROM phone_web_occurrences WHERE phone_number_id=$1 AND is_active=TRUE ORDER BY source_confidence DESC,last_seen_at DESC LIMIT $2`,phoneID,limit);if err!=nil{return summary,err};defer items.Close()
 for items.Next(){var item WebOccurrence;if err:=items.Scan(&item.ID,&item.Domain,&item.URL,&item.PageTitle,&item.Category,&item.ContextSnippet,&item.SourceConfidence,&item.PublishedAt,&item.DetectedAt,&item.LastSeenAt,&item.LastCheckedAt);err!=nil{return summary,err};summary.Occurrences=append(summary.Occurrences,item)}
 return summary,items.Err()
}

func (r Repository) QueueFootprintScan(ctx context.Context, phoneID string) (string,error) {
 var id string
 err:=r.DB.QueryRowContext(ctx,`INSERT INTO phone_web_scan_jobs(phone_number_id) SELECT $1 WHERE NOT EXISTS(SELECT 1 FROM phone_web_scan_jobs WHERE phone_number_id=$1 AND status IN ('pending','running')) RETURNING id::text`,phoneID).Scan(&id)
 if errors.Is(err,sql.ErrNoRows){err=r.DB.QueryRowContext(ctx,`SELECT id::text FROM phone_web_scan_jobs WHERE phone_number_id=$1 AND status IN ('pending','running') ORDER BY requested_at DESC LIMIT 1`,phoneID).Scan(&id)}
 return id,err
}

func (r Repository) SitemapShard(ctx context.Context, shard, shards, limit int) ([]SitemapNumber,error){
 if shards<1||shards>4096||shard<0||shard>=shards{return []SitemapNumber{},nil};if limit<1||limit>50000{limit=50000}
 rows,err:=r.DB.QueryContext(ctx,`SELECT e164,updated_at FROM phone_numbers WHERE seo_status IN ('indexable','indexed') AND mod(abs(hashtext(id::text)::bigint),$2)=$1 ORDER BY id LIMIT $3`,shard,shards,limit);if err!=nil{return nil,err};defer rows.Close()
 items:=make([]SitemapNumber,0);for rows.Next(){var x SitemapNumber;if err:=rows.Scan(&x.E164,&x.UpdatedAt);err!=nil{return nil,err};items=append(items,x)};return items,rows.Err()
}
