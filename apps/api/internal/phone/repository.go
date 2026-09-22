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
