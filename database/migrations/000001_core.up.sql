CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE verification_status AS ENUM (
  'unverified',
  'pending',
  'verified',
  'disputed'
);

CREATE TYPE seo_status AS ENUM (
  'discovered',
  'noindex',
  'indexable',
  'indexed',
  'review'
);

CREATE TABLE phone_numbers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  country_code CHAR(2) NOT NULL,
  calling_code VARCHAR(8) NOT NULL,
  national_number VARCHAR(32) NOT NULL,
  e164 VARCHAR(32) NOT NULL UNIQUE,
  number_type VARCHAR(32),
  carrier_id UUID,
  region_id UUID,
  verification_status verification_status NOT NULL DEFAULT 'unverified',
  seo_status seo_status NOT NULL DEFAULT 'discovered',
  spam_score NUMERIC(5,4) NOT NULL DEFAULT 0 CHECK (spam_score >= 0 AND spam_score <= 1),
  report_count BIGINT NOT NULL DEFAULT 0 CHECK (report_count >= 0),
  search_count BIGINT NOT NULL DEFAULT 0 CHECK (search_count >= 0),
  data_quality_score NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK (data_quality_score >= 0 AND data_quality_score <= 100),
  first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(country_code, national_number)
);

CREATE INDEX idx_phone_numbers_country_national
  ON phone_numbers(country_code, national_number);

CREATE INDEX idx_phone_numbers_seo_status
  ON phone_numbers(seo_status)
  WHERE seo_status IN ('indexable', 'indexed');
