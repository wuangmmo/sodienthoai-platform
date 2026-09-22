CREATE TABLE phone_web_occurrences (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 phone_number_id UUID NOT NULL REFERENCES phone_numbers(id) ON DELETE CASCADE,
 domain VARCHAR(255) NOT NULL,
 url TEXT NOT NULL,
 normalized_url TEXT NOT NULL,
 page_title TEXT,
 category VARCHAR(64) NOT NULL DEFAULT 'other',
 context_snippet TEXT,
 source_confidence NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK(source_confidence BETWEEN 0 AND 100),
 cluster_key VARCHAR(128),
 published_at TIMESTAMPTZ,
 detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 last_checked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 is_active BOOLEAN NOT NULL DEFAULT TRUE,
 UNIQUE(phone_number_id, normalized_url)
);
CREATE INDEX idx_phone_web_occurrences_phone ON phone_web_occurrences(phone_number_id,last_seen_at DESC);
CREATE INDEX idx_phone_web_occurrences_domain ON phone_web_occurrences(phone_number_id,domain);
CREATE INDEX idx_phone_web_occurrences_category ON phone_web_occurrences(phone_number_id,category);

CREATE TABLE phone_web_scan_jobs (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 phone_number_id UUID NOT NULL REFERENCES phone_numbers(id) ON DELETE CASCADE,
 status VARCHAR(24) NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','running','success','failed')),
 requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 started_at TIMESTAMPTZ,
 completed_at TIMESTAMPTZ,
 result_count INTEGER NOT NULL DEFAULT 0,
 last_error TEXT
);
CREATE INDEX idx_phone_web_scan_jobs_pending ON phone_web_scan_jobs(status,requested_at);
