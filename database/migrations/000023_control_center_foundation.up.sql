CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','disabled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    domain TEXT NOT NULL UNIQUE,
    site_type TEXT NOT NULL CHECK (site_type IN ('identity_platform','local_directory','web_archive','service_site','connected_platform')),
    management_mode TEXT NOT NULL DEFAULT 'managed' CHECK (management_mode IN ('managed','connected')),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','disabled')),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, slug)
);

CREATE TABLE IF NOT EXISTS control_permissions (
    key TEXT PRIMARY KEY,
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS control_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, slug)
);

CREATE TABLE IF NOT EXISTS control_role_permissions (
    role_id UUID NOT NULL REFERENCES control_roles(id) ON DELETE CASCADE,
    permission_key TEXT NOT NULL REFERENCES control_permissions(key) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_key)
);

CREATE TABLE IF NOT EXISTS control_admin_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_user_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES control_roles(id) ON DELETE CASCADE,
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    site_id UUID REFERENCES sites(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (site_id IS NULL OR organization_id IS NOT NULL),
    UNIQUE (admin_user_id, role_id, organization_id, site_id)
);

CREATE TABLE IF NOT EXISTS control_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    admin_user_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    site_id UUID REFERENCES sites(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    before_data JSONB,
    after_data JSONB,
    ip_address INET,
    request_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_control_assignments_admin ON control_admin_assignments(admin_user_id);
CREATE INDEX IF NOT EXISTS idx_control_assignments_site ON control_admin_assignments(site_id);
CREATE INDEX IF NOT EXISTS idx_control_audit_created ON control_audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_control_audit_admin ON control_audit_logs(admin_user_id, created_at DESC);

INSERT INTO organizations (slug, name)
VALUES ('sodienthoai-ecosystem', 'Sodienthoai Ecosystem')
ON CONFLICT (slug) DO NOTHING;

WITH org AS (SELECT id FROM organizations WHERE slug='sodienthoai-ecosystem')
INSERT INTO sites (organization_id, slug, name, domain, site_type, management_mode)
SELECT org.id, x.slug, x.name, x.domain, x.site_type, x.management_mode
FROM org
CROSS JOIN (VALUES
  ('sodienthoai-com','Sodienthoai.com','sodienthoai.com','identity_platform','managed'),
  ('sodienthoai-vn','Sodienthoai.vn','sodienthoai.vn','local_directory','managed'),
  ('sodienthoai-org','Sodienthoai.org','sodienthoai.org','web_archive','managed'),
  ('mapsviet','MapsViet','mapsviet.com','connected_platform','connected')
) AS x(slug,name,domain,site_type,management_mode)
ON CONFLICT (domain) DO NOTHING;

INSERT INTO control_permissions (key, description) VALUES
('dashboard.view','View Control Center dashboard'),
('phones.view','View phone intelligence'),
('phones.manage','Manage phone identities and metadata'),
('reports.moderate','Moderate reports and community content'),
('claims.review','Review phone claims and verification evidence'),
('directory.manage','Manage local directory data'),
('businesses.manage','Manage businesses and reviews'),
('hotlines.view','View hotline registry'),
('hotlines.manage','Manage hotline assignments'),
('sites.view','View sites and connected platforms'),
('sites.manage','Manage sites and integrations'),
('users.view','View user accounts and activity'),
('users.manage','Manage user accounts'),
('seo.manage','Manage SEO and indexing'),
('data.manage','Manage imports and data quality'),
('operations.view','View jobs and system health'),
('admins.manage','Manage administrators'),
('roles.manage','Manage roles, permissions and scopes'),
('audit.view','View audit logs'),
('system.manage','Manage system settings')
ON CONFLICT (key) DO NOTHING;

WITH org AS (SELECT id FROM organizations WHERE slug='sodienthoai-ecosystem')
INSERT INTO control_roles (organization_id, slug, name, description, is_system)
SELECT org.id, x.slug, x.name, x.description, TRUE
FROM org
CROSS JOIN (VALUES
 ('super-admin','Super Admin','Full organization control'),
 ('admin','Administrator','General platform administration'),
 ('moderator','Moderator','Reports, comments and verification moderation'),
 ('data-operator','Data Operator','Imports, directory and data quality operations'),
 ('seo-content','SEO / Content','SEO, indexing and content operations'),
 ('support','Support','User support and read access'),
 ('analyst','Analyst','Read-only analytics and operational visibility'),
 ('read-only','Read Only','Read-only Control Center access')
) AS x(slug,name,description)
ON CONFLICT (organization_id,slug) DO NOTHING;

WITH roles AS (
 SELECT r.id,r.slug FROM control_roles r
 JOIN organizations o ON o.id=r.organization_id
 WHERE o.slug='sodienthoai-ecosystem'
)
INSERT INTO control_role_permissions(role_id,permission_key)
SELECT roles.id,p.key FROM roles CROSS JOIN control_permissions p
WHERE roles.slug='super-admin'
ON CONFLICT DO NOTHING;

WITH org AS (SELECT id FROM organizations WHERE slug='sodienthoai-ecosystem'),
role AS (
 SELECT r.id FROM control_roles r JOIN org ON r.organization_id=org.id WHERE r.slug='super-admin'
)
INSERT INTO control_admin_assignments(admin_user_id,role_id,organization_id)
SELECT a.id,role.id,org.id FROM admin_users a CROSS JOIN role CROSS JOIN org
WHERE a.role='super-admin'
ON CONFLICT DO NOTHING;
