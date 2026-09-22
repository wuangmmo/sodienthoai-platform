CREATE INDEX idx_phone_reports_approved_phone ON phone_reports(phone_number_id) WHERE status='approved';

CREATE OR REPLACE FUNCTION refresh_phone_reputation(target_phone UUID)
RETURNS VOID LANGUAGE plpgsql AS $$
DECLARE approved_count BIGINT;
BEGIN
 SELECT COUNT(*) INTO approved_count FROM phone_reports WHERE phone_number_id=target_phone AND status='approved';
 UPDATE phone_numbers
 SET report_count=approved_count,
     spam_score=ROUND((1-EXP(-approved_count::numeric/5))::numeric,4),
     updated_at=NOW()
 WHERE id=target_phone;
END;
$$;
