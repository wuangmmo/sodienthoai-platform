CREATE TABLE seo_index_queue (
 phone_number_id UUID PRIMARY KEY REFERENCES phone_numbers(id) ON DELETE CASCADE,
 action VARCHAR(16) NOT NULL CHECK(action IN ('index','remove')),
 status VARCHAR(16) NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','processing','done','failed')),
 attempts INTEGER NOT NULL DEFAULT 0,
 available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 last_error TEXT,
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_seo_index_queue_pending ON seo_index_queue(available_at,phone_number_id) WHERE status IN ('pending','failed');
CREATE OR REPLACE FUNCTION queue_phone_seo_change() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF NEW.seo_status IS DISTINCT FROM OLD.seo_status THEN
  INSERT INTO seo_index_queue(phone_number_id,action,status,available_at,updated_at)
  VALUES(NEW.id,CASE WHEN NEW.seo_status IN ('indexable','indexed') THEN 'index' ELSE 'remove' END,'pending',NOW(),NOW())
  ON CONFLICT(phone_number_id) DO UPDATE SET action=EXCLUDED.action,status='pending',available_at=NOW(),last_error=NULL,updated_at=NOW();
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER trg_phone_seo_queue AFTER UPDATE OF seo_status ON phone_numbers FOR EACH ROW EXECUTE FUNCTION queue_phone_seo_change();
