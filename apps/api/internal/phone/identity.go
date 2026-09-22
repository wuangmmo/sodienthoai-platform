package phone

type Identity struct {
	ID              string  `json:"id"`
	Kind            string  `json:"kind"`
	DisplayName     string  `json:"display_name"`
	Description     *string `json:"description,omitempty"`
	WebsiteURL      *string `json:"website_url,omitempty"`
	AddressText     *string `json:"address_text,omitempty"`
	SourceLabel     *string `json:"source_label,omitempty"`
	IsPrimary       bool    `json:"is_primary"`
	ConfidenceScore float64 `json:"confidence_score"`
}
