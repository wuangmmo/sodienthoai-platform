CREATE TABLE phone_import_batches (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 source_name VARCHAR(160) NOT NULL,
 status VARCHAR(32) NOT NULL DEFAULT 'pending',
 total_rows BIGINT NOT NULL DEFAULT 0,
 accepted_rows BIGINT NOT NULL DEFAULT 0,
 rejected_rows BIGINT NOT NULL DEFAULT 0,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 completed_at TIMESTAMPTZ
);
CREATE INDEX idx_phone_numbers_search_rank ON phone_numbers(search_count DESC,last_seen_at DESC);
CREATE INDEX idx_phone_numbers_country_search ON phone_numbers(country_code,search_count DESC);
