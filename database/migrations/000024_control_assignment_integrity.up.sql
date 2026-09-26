CREATE UNIQUE INDEX IF NOT EXISTS uq_control_assignment_org_scope ON control_admin_assignments(admin_user_id,role_id,organization_id) WHERE site_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_control_assignment_site_scope ON control_admin_assignments(admin_user_id,role_id,organization_id,site_id) WHERE site_id IS NOT NULL;
