CREATE TABLE phone_sources (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  phone_number_id UUID NOT NULL REFERENCES phone_numbers(id) ON DELETE CASCADE,
  source_type VARCHAR(32) NOT NULL,
  source_ref TEXT,
  confidence NUMERIC(5,4) NOT NULL DEFAULT 0.5 CHECK (confidence >= 0 AND confidence <= 1),
  observed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_phone_sources_phone_number_id
  ON phone_sources(phone_number_id);

CREATE INDEX idx_phone_sources_type
  ON phone_sources(source_type);
