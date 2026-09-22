package phone

import "time"

type ProfileSignals struct {
	IdentityConfidence    float64    `json:"identity_confidence"`
	SourceDiversity       int64      `json:"source_diversity"`
	FootprintSourceCount  int64      `json:"footprint_source_count"`
	FootprintDomainCount  int64      `json:"footprint_domain_count"`
	FootprintCategoryCount int64     `json:"footprint_category_count"`
	FirstDetectedAt       *time.Time `json:"first_detected_at,omitempty"`
	LastDetectedAt        *time.Time `json:"last_detected_at,omitempty"`
}

type Profile struct {
	Number          Number         `json:"number"`
	PrimaryIdentity *Identity      `json:"primary_identity,omitempty"`
	Identities      []Identity     `json:"identities"`
	Signals         ProfileSignals `json:"signals"`
	Identified      bool           `json:"identified"`
	Disputed        bool           `json:"disputed"`
}
