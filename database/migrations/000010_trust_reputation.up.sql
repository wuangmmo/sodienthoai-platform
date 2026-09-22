CREATE TABLE phone_verification_evidence (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 claim_id UUID NOT NULL REFERENCES phone_claims(id) ON DELETE CASCADE,
 evidence_type VARCHAR(32) NOT NULL CHECK(evidence_type IN ('website','document','dns','phone','other')),
 evidence_value TEXT NOT NULL,
 status moderation_status NOT NULL DEFAULT 'pending',
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 reviewed_at TIMESTAMPTZ
);
CREATE INDEX idx_verification_evidence_claim ON phone_verification_evidence(claim_id,status);
CREATE TABLE phone_reputation_events (
 id BIGSERIAL PRIMARY KEY,
 phone_number_id UUID NOT NULL REFERENCES phone_numbers(id) ON DELETE CASCADE,
 event_type VARCHAR(40) NOT NULL,
 score_delta NUMERIC(6,2) NOT NULL DEFAULT 0,
 source VARCHAR(80),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_reputation_events_phone ON phone_reputation_events(phone_number_id,created_at DESC);
ALTER TABLE phone_numbers ADD COLUMN trust_score NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK(trust_score BETWEEN 0 AND 100);
ALTER TABLE phone_numbers ADD COLUMN reputation_updated_at TIMESTAMPTZ;
