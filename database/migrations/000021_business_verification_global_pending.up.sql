-- V42: enforce one active verification workflow per business at the database boundary.
-- The service also serializes requests with SELECT ... FOR UPDATE; this index protects
-- against concurrent/direct writers that bypass that code path.
DROP INDEX IF EXISTS idx_business_verification_one_pending;
CREATE UNIQUE INDEX idx_business_verification_one_pending
 ON business_verification_requests(business_id)
 WHERE status='pending';
