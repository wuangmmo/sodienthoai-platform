package phone

import "time"

type WebOccurrence struct {
 ID string `json:"id"`
 Domain string `json:"domain"`
 URL string `json:"url"`
 PageTitle *string `json:"page_title,omitempty"`
 Category string `json:"category"`
 ContextSnippet *string `json:"context_snippet,omitempty"`
 SourceConfidence float64 `json:"source_confidence"`
 PublishedAt *time.Time `json:"published_at,omitempty"`
 DetectedAt time.Time `json:"detected_at"`
 LastSeenAt time.Time `json:"last_seen_at"`
 LastCheckedAt time.Time `json:"last_checked_at"`
}

type FootprintSummary struct {
 OccurrenceCount int64 `json:"occurrence_count"`
 IndependentDomainCount int64 `json:"independent_domain_count"`
 CategoryCount int64 `json:"category_count"`
 FirstDetectedAt *time.Time `json:"first_detected_at,omitempty"`
 LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
 Categories map[string]int64 `json:"categories"`
 Occurrences []WebOccurrence `json:"occurrences"`
}
