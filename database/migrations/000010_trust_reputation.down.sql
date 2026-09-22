ALTER TABLE phone_numbers DROP COLUMN IF EXISTS reputation_updated_at;
ALTER TABLE phone_numbers DROP COLUMN IF EXISTS trust_score;
DROP TABLE IF EXISTS phone_reputation_events;
DROP TABLE IF EXISTS phone_verification_evidence;
