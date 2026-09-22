ALTER TABLE phone_claims DROP COLUMN IF EXISTS moderation_note;
ALTER TABLE phone_reports DROP COLUMN IF EXISTS moderation_note;
ALTER TABLE phone_reports DROP COLUMN IF EXISTS reviewed_at;
DROP INDEX IF EXISTS idx_phone_reports_reporter_pending_unique;
