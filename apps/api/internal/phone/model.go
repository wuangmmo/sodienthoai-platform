package phone

import "time"

type Number struct {
	ID                 string    `json:"id"`
	CountryCode        string    `json:"country_code"`
	CallingCode        string    `json:"calling_code"`
	NationalNumber     string    `json:"national_number"`
	E164               string    `json:"e164"`
	NumberType         *string   `json:"number_type,omitempty"`
	VerificationStatus string    `json:"verification_status"`
	SEOStatus          string    `json:"seo_status"`
	SpamScore          float64   `json:"spam_score"`
	ReportCount        int64     `json:"report_count"`
	SearchCount        int64     `json:"search_count"`
	DataQualityScore   float64   `json:"data_quality_score"`
	FirstSeenAt        time.Time `json:"first_seen_at"`
	LastSeenAt         time.Time `json:"last_seen_at"`
	ReputationLabel    string    `json:"reputation_label"`
}
