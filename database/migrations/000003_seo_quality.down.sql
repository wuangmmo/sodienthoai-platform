DROP INDEX IF EXISTS idx_phone_numbers_sitemap;
ALTER TABLE phone_numbers DROP COLUMN IF EXISTS seo_reason;
ALTER TABLE phone_numbers DROP COLUMN IF EXISTS seo_evaluated_at;
