INSERT INTO phone_numbers (
  country_code,
  calling_code,
  national_number,
  e164,
  number_type,
  verification_status,
  seo_status,
  spam_score,
  report_count,
  search_count,
  data_quality_score
) VALUES (
  'VN',
  '84',
  '705899899',
  '+84705899899',
  'mobile',
  'verified',
  'indexable',
  0,
  0,
  1,
  95
)
ON CONFLICT (e164) DO NOTHING;

INSERT INTO phone_sources (
  phone_number_id,
  source_type,
  source_ref,
  confidence
)
SELECT id, 'ci_seed', 'database/testdata/seed.sql', 1
FROM phone_numbers
WHERE e164 = '+84705899899'
  AND NOT EXISTS (
    SELECT 1 FROM phone_sources
    WHERE phone_number_id = phone_numbers.id
      AND source_type = 'ci_seed'
  );
