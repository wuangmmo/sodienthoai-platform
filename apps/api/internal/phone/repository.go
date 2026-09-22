package phone

import (
	"context"
	"database/sql"
	"errors"
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
