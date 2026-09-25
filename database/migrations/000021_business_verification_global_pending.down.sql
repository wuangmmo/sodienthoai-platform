DROP INDEX IF EXISTS idx_business_verification_one_pending;
CREATE UNIQUE INDEX idx_business_verification_one_pending
 ON business_verification_requests(business_id,user_id)
 WHERE status='pending';
