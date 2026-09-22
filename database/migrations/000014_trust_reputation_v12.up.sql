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

CREATE OR REPLACE FUNCTION refresh_phone_reputation(target_phone UUID)
RETURNS VOID LANGUAGE plpgsql AS $$
DECLARE approved_count BIGINT;
DECLARE calculated_spam NUMERIC(5,4);
DECLARE label VARCHAR(24);
DECLARE disputed BOOLEAN;
BEGIN
 SELECT COUNT(*) INTO approved_count FROM phone_reports WHERE phone_number_id=target_phone AND status='approved';
 calculated_spam:=ROUND((1-EXP(-approved_count::numeric/5))::numeric,4);
 SELECT EXISTS(SELECT 1 FROM phone_identity_disputes WHERE phone_number_id=target_phone AND status='open') INTO disputed;
 IF disputed THEN label:='disputed';
 ELSIF approved_count>=5 OR calculated_spam>=0.65 THEN label:='high_risk';
 ELSIF approved_count>=1 THEN label:='caution';
 ELSE label:='clear';
 END IF;
 UPDATE phone_numbers SET report_count=approved_count,spam_score=calculated_spam,reputation_label=label,updated_at=NOW() WHERE id=target_phone;
END;
$$;
