DROP INDEX IF EXISTS idx_phone_numbers_reputation_label;
ALTER TABLE phone_numbers DROP COLUMN IF EXISTS reputation_label;
DROP TABLE IF EXISTS phone_report_fingerprints;
