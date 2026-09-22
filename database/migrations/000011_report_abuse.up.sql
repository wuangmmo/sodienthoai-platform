CREATE UNIQUE INDEX idx_phone_reports_reporter_pending_unique ON phone_reports(phone_number_id,reporter_hash,reason) WHERE status='pending' AND reporter_hash IS NOT NULL;
ALTER TABLE phone_reports ADD COLUMN reviewed_at TIMESTAMPTZ;
ALTER TABLE phone_reports ADD COLUMN moderation_note TEXT;
ALTER TABLE phone_claims ADD COLUMN moderation_note TEXT;
