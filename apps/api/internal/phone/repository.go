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
       report_count, search_count, data_quality_score::float8, first_seen_at, last_seen_at
FROM phone_numbers WHERE e164 = $1
`
	var n Number
	err := r.DB.QueryRowContext(ctx,q,e164).Scan(
		&n.ID,&n.CountryCode,&n.CallingCode,&n.NationalNumber,&n.E164,&n.NumberType,
		&n.VerificationStatus,&n.SEOStatus,&n.SpamScore,&n.ReportCount,&n.SearchCount,
		&n.DataQualityScore,&n.FirstSeenAt,&n.LastSeenAt,
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

func (r Repository) SitemapPage(ctx context.Context, limit, offset int) ([]SitemapNumber, error) {
	if limit < 1 || limit > 50000 { limit = 50000 }
	if offset < 0 { offset = 0 }
	rows, err := r.DB.QueryContext(ctx, `
SELECT e164, updated_at
FROM phone_numbers
WHERE seo_status IN ('indexable','indexed')
ORDER BY updated_at DESC, e164
LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, err }
	defer rows.Close()
	items := make([]SitemapNumber,0)
	for rows.Next(){var item SitemapNumber;if err:=rows.Scan(&item.E164,&item.UpdatedAt);err!=nil{return nil,err};items=append(items,item)}
	return items,rows.Err()
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
