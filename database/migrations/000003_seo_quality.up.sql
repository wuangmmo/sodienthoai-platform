ALTER TABLE phone_numbers
  ADD COLUMN seo_evaluated_at TIMESTAMPTZ,
  ADD COLUMN seo_reason VARCHAR(64);

CREATE INDEX idx_phone_numbers_sitemap
  ON phone_numbers(updated_at DESC, e164)
  WHERE seo_status IN ('indexable','indexed');
