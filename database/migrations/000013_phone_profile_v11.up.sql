CREATE TYPE identity_match_status AS ENUM ('match','possible_match','mismatch','unknown','disputed');

CREATE TABLE phone_profile_signals (
  phone_number_id UUID PRIMARY KEY REFERENCES phone_numbers(id) ON DELETE CASCADE,
  identity_confidence NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK (identity_confidence BETWEEN 0 AND 100),
  source_diversity INTEGER NOT NULL DEFAULT 0 CHECK (source_diversity >= 0),
  footprint_source_count INTEGER NOT NULL DEFAULT 0 CHECK (footprint_source_count >= 0),
  footprint_domain_count INTEGER NOT NULL DEFAULT 0 CHECK (footprint_domain_count >= 0),
  footprint_category_count INTEGER NOT NULL DEFAULT 0 CHECK (footprint_category_count >= 0),
  footprint_first_detected_at TIMESTAMPTZ,
  footprint_last_detected_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE phone_identity_suggestions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  phone_number_id UUID NOT NULL REFERENCES phone_numbers(id) ON DELETE CASCADE,
  suggested_kind identity_kind NOT NULL DEFAULT 'unknown',
  suggested_display_name VARCHAR(255) NOT NULL,
  reason TEXT,
  evidence_url TEXT,
  status VARCHAR(24) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected','withdrawn')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  reviewed_at TIMESTAMPTZ
);
CREATE INDEX idx_phone_identity_suggestions_review ON phone_identity_suggestions(status, created_at);

CREATE TABLE phone_identity_disputes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  phone_number_id UUID NOT NULL REFERENCES phone_numbers(id) ON DELETE CASCADE,
  identity_id UUID REFERENCES phone_identities(id) ON DELETE SET NULL,
  reason TEXT NOT NULL,
  status VARCHAR(24) NOT NULL DEFAULT 'open' CHECK (status IN ('open','resolved','rejected')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  resolved_at TIMESTAMPTZ
);
CREATE INDEX idx_phone_identity_disputes_open ON phone_identity_disputes(phone_number_id, status);
