CREATE INDEX IF NOT EXISTS idx_phone_numbers_sitemap ON phone_numbers(id) WHERE seo_status IN ('indexable','indexed');
CREATE INDEX IF NOT EXISTS idx_phone_numbers_seo_country ON phone_numbers(country_code,id) WHERE seo_status IN ('indexable','indexed');
