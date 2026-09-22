CREATE TYPE report_reason AS ENUM ('spam','scam','telemarketing','harassment','wrong_identity','other');
CREATE TYPE moderation_status AS ENUM ('pending','approved','rejected');
CREATE TYPE claim_status AS ENUM ('pending','verified','rejected');

CREATE TABLE phone_reports (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 phone_number_id UUID REFERENCES phone_numbers(id) ON DELETE CASCADE,
 e164 VARCHAR(32) NOT NULL,
 reason report_reason NOT NULL,
 comment TEXT,
 reporter_hash VARCHAR(128),
 status moderation_status NOT NULL DEFAULT 'pending',
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 reviewed_at TIMESTAMPTZ
);
CREATE INDEX idx_phone_reports_phone_status ON phone_reports(phone_number_id,status);
CREATE INDEX idx_phone_reports_e164_created ON phone_reports(e164,created_at DESC);

CREATE TABLE phone_claims (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 phone_number_id UUID NOT NULL REFERENCES phone_numbers(id) ON DELETE CASCADE,
 claimant_name VARCHAR(255) NOT NULL,
 claimant_type identity_kind NOT NULL DEFAULT 'unknown',
 contact_email VARCHAR(320),
 evidence_note TEXT,
 status claim_status NOT NULL DEFAULT 'pending',
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 reviewed_at TIMESTAMPTZ
);
CREATE INDEX idx_phone_claims_phone_status ON phone_claims(phone_number_id,status);
