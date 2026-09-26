DROP TRIGGER IF EXISTS trg_control_assignment_site_org ON control_admin_assignments;
DROP FUNCTION IF EXISTS enforce_control_assignment_site_org();
DROP INDEX IF EXISTS uq_control_assignment_site_scope;
DROP INDEX IF EXISTS uq_control_assignment_org_scope;
