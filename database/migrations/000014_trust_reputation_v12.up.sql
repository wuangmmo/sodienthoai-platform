CREATE TABLE phone_report_fingerprints (
  phone_number_id UUID NOT NULL REFERENCES phone_numbers(id) ON DELETE CASCADE,
  reporter_hash VARCHAR(128) NOT NULL,
  reason VARCHAR(64) NOT NULL,
  last_reported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY(phone_number_id, reporter_hash, reason)
);
CREATE INDEX idx_phone_report_fingerprints_recent ON phone_report_fingerprints(last_reported_at DESC);

ALTER TABLE phone_numbers
  ADD COLUMN IF NOT EXISTS reputation_label VARCHAR(24) NOT NULL DEFAULT 'unknown'
  CHECK (reputation_label IN ('unknown','clear','caution','high_risk','disputed'));

CREATE INDEX idx_phone_numbers_reputation_label ON phone_numbers(reputation_label);
