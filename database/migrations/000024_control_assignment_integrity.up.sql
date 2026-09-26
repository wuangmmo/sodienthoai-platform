CREATE UNIQUE INDEX IF NOT EXISTS uq_control_assignment_org_scope ON control_admin_assignments(admin_user_id,role_id,organization_id) WHERE site_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_control_assignment_site_scope ON control_admin_assignments(admin_user_id,role_id,organization_id,site_id) WHERE site_id IS NOT NULL;

CREATE OR REPLACE FUNCTION enforce_control_assignment_site_org() RETURNS trigger AS $$
BEGIN
  IF NEW.site_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM sites s WHERE s.id=NEW.site_id AND s.organization_id=NEW.organization_id) THEN
    RAISE EXCEPTION 'control assignment site organization mismatch';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_control_assignment_site_org ON control_admin_assignments;
CREATE TRIGGER trg_control_assignment_site_org BEFORE INSERT OR UPDATE OF organization_id,site_id ON control_admin_assignments FOR EACH ROW EXECUTE FUNCTION enforce_control_assignment_site_org();
