CREATE TABLE admin_audit_log (
 id BIGSERIAL PRIMARY KEY,
 actor VARCHAR(160) NOT NULL,
 action VARCHAR(80) NOT NULL,
 resource_type VARCHAR(80) NOT NULL,
 resource_id VARCHAR(160),
 metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_admin_audit_log_created ON admin_audit_log(created_at DESC);
CREATE INDEX idx_admin_audit_resource ON admin_audit_log(resource_type,resource_id,created_at DESC);
CREATE TABLE operational_jobs (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 job_type VARCHAR(80) NOT NULL,
 status VARCHAR(24) NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','running','success','failed')),
 payload JSONB NOT NULL DEFAULT '{}'::jsonb,
 attempts INTEGER NOT NULL DEFAULT 0,
 last_error TEXT,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 started_at TIMESTAMPTZ,
 completed_at TIMESTAMPTZ
);
CREATE INDEX idx_operational_jobs_pending ON operational_jobs(status,created_at) WHERE status IN ('pending','failed');
