CREATE TYPE identity_kind AS ENUM ('person','business','organization','service','unknown');

CREATE TABLE phone_identities (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  phone_number_id UUID NOT NULL REFERENCES phone_numbers(id) ON DELETE CASCADE,
  kind identity_kind NOT NULL DEFAULT 'unknown',
  display_name VARCHAR(255) NOT NULL,
  description TEXT,
  website_url TEXT,
  address_text TEXT,
  source_label VARCHAR(120),
  is_primary BOOLEAN NOT NULL DEFAULT FALSE,
  is_public BOOLEAN NOT NULL DEFAULT TRUE,
  confidence_score NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK (confidence_score >= 0 AND confidence_score <= 100),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_phone_identities_phone ON phone_identities(phone_number_id);
CREATE UNIQUE INDEX idx_phone_identity_primary ON phone_identities(phone_number_id) WHERE is_primary = TRUE;

CREATE TABLE phone_lookup_events (
  id BIGSERIAL PRIMARY KEY,
  phone_number_id UUID REFERENCES phone_numbers(id) ON DELETE SET NULL,
  e164 VARCHAR(32) NOT NULL,
  country_code CHAR(2),
  found BOOLEAN NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_phone_lookup_events_e164_created ON phone_lookup_events(e164, created_at DESC);
