package httpserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net"
	"regexp"
	"strings"
	"time"
)

type ControlHandler struct {
	DB   *sql.DB
	Auth *AdminAuthenticator
}

type controlAccess struct {
	Admin       adminDTO
	Permissions map[string]bool
	SiteIDs     []string
}

func (h ControlHandler) access(r *http.Request) (controlAccess, bool) {
	var out controlAccess
	claims, ok := h.Auth.authenticate(r)
	if !ok { return out, false }

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	err := h.DB.QueryRowContext(ctx, `SELECT id::text,name,email,role,status,created_at,last_login_at
		FROM admin_users WHERE id=$1::uuid LIMIT 1`, claims.Sub).
		Scan(&out.Admin.ID,&out.Admin.Name,&out.Admin.Email,&out.Admin.Role,&out.Admin.Status,&out.Admin.CreatedAt,&out.Admin.LastLoginAt)
	if err != nil || out.Admin.Status != "active" { return controlAccess{}, false }

	out.Permissions = map[string]bool{}
	rows, err := h.DB.QueryContext(ctx, `
		SELECT DISTINCT rp.permission_key
		FROM control_admin_assignments aa
		JOIN control_role_permissions rp ON rp.role_id=aa.role_id
		WHERE aa.admin_user_id=$1::uuid`, claims.Sub)
	if err != nil { return controlAccess{}, false }
	defer rows.Close()
	for rows.Next() {
		var key string
		if rows.Scan(&key) == nil { out.Permissions[key]=true }
	}
	if rows.Err()!=nil { return controlAccess{}, false }

	siteRows, err := h.DB.QueryContext(ctx, `
		SELECT DISTINCT s.id::text
		FROM sites s
		JOIN control_admin_assignments aa ON aa.admin_user_id=$1::uuid
		WHERE aa.site_id=s.id OR (aa.site_id IS NULL AND aa.organization_id=s.organization_id)`, claims.Sub)
	if err != nil { return controlAccess{}, false }
	defer siteRows.Close()
	for siteRows.Next() {
		var id string
		if siteRows.Scan(&id)==nil { out.SiteIDs=append(out.SiteIDs,id) }
	}
	return out, siteRows.Err()==nil
}

func (h ControlHandler) Me(w http.ResponseWriter, r *http.Request) {
	a, ok := h.access(r)
	if !ok { writeJSON(w,http.StatusUnauthorized,map[string]string{"error":"unauthorized"}); return }
	perms:=make([]string,0,len(a.Permissions))
	for p:=range a.Permissions { perms=append(perms,p) }
	writeJSON(w,http.StatusOK,map[string]any{"admin":a.Admin,"permissions":perms,"siteIds":a.SiteIDs})
}

func (h ControlHandler) Sites(w http.ResponseWriter, r *http.Request) {
	a, ok := h.access(r)
	if !ok { writeJSON(w,http.StatusUnauthorized,map[string]string{"error":"unauthorized"}); return }
	if !a.Permissions["sites.view"] && !a.Permissions["sites.manage"] {
		writeJSON(w,http.StatusForbidden,map[string]string{"error":"forbidden"}); return
	}
	ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second); defer cancel()
	rows,err:=h.DB.QueryContext(ctx,`
		SELECT s.id::text,s.slug,s.name,s.domain,s.site_type,s.management_mode,s.status,o.id::text,o.name
		FROM sites s JOIN organizations o ON o.id=s.organization_id
		WHERE s.id::text = ANY($1::text[])
		ORDER BY s.name`, pqTextArray(a.SiteIDs))
	if err!=nil { writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"}); return }
	defer rows.Close()
	type item struct { ID,Slug,Name,Domain,SiteType,ManagementMode,Status,OrganizationID,OrganizationName string }
	items:=[]item{}
	for rows.Next(){ var x item; if rows.Scan(&x.ID,&x.Slug,&x.Name,&x.Domain,&x.SiteType,&x.ManagementMode,&x.Status,&x.OrganizationID,&x.OrganizationName)==nil { items=append(items,x) } }
	writeJSON(w,http.StatusOK,map[string]any{"sites":items})
}

// pqTextArray returns a PostgreSQL text[] literal without adding a driver dependency.
// Values are UUIDs generated/read by this service and therefore contain no array metacharacters.
func pqTextArray(values []string) string {
	if len(values)==0 { return "{}" }
	out:="{"
	for i,v:=range values { if i>0 { out+="," }; out+=v }
	return out+"}"
}

func (h ControlHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	a,ok:=h.access(r)
	if !ok { writeJSON(w,http.StatusUnauthorized,map[string]string{"error":"unauthorized"}); return }
	if !a.Permissions["dashboard.view"] { writeJSON(w,http.StatusForbidden,map[string]string{"error":"forbidden"}); return }
	ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second); defer cancel()
	var phones,reports,claims,sites int64
	if err:=h.DB.QueryRowContext(ctx,`SELECT
		(SELECT count(*) FROM phone_numbers),
		(SELECT count(*) FROM phone_reports),
		(SELECT count(*) FROM phone_claims),
		(SELECT count(*) FROM sites WHERE id::text = ANY($1::text[]))`,pqTextArray(a.SiteIDs)).Scan(&phones,&reports,&claims,&sites); err!=nil {
		writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"}); return
	}
	writeJSON(w,http.StatusOK,map[string]any{"phones":phones,"reports":reports,"claims":claims,"sites":sites})
}


func (h ControlHandler) Roles(w http.ResponseWriter, r *http.Request) {
	a,ok:=h.access(r)
	if !ok { writeJSON(w,http.StatusUnauthorized,map[string]string{"error":"unauthorized"}); return }
	if !a.Permissions["roles.manage"] { writeJSON(w,http.StatusForbidden,map[string]string{"error":"forbidden"}); return }
	ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second); defer cancel()
	rows,err:=h.DB.QueryContext(ctx,`
		SELECT r.id::text,r.slug,r.name,r.description,r.is_system,
		       COALESCE(array_agg(rp.permission_key ORDER BY rp.permission_key) FILTER (WHERE rp.permission_key IS NOT NULL),'{}')
		FROM control_roles r
		LEFT JOIN control_role_permissions rp ON rp.role_id=r.id
		WHERE r.organization_id IN (
		  SELECT DISTINCT organization_id FROM control_admin_assignments WHERE admin_user_id=$1::uuid AND organization_id IS NOT NULL
		)
		GROUP BY r.id,r.slug,r.name,r.description,r.is_system
		ORDER BY r.name`,a.Admin.ID)
	if err!=nil { writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"}); return }
	defer rows.Close()
	type role struct { ID,Slug,Name,Description string; IsSystem bool; Permissions []string }
	items:=[]role{}
	for rows.Next(){var x role;var raw string;if rows.Scan(&x.ID,&x.Slug,&x.Name,&x.Description,&x.IsSystem,&raw)==nil{x.Permissions=parsePGTextArray(raw);items=append(items,x)}}
	writeJSON(w,http.StatusOK,map[string]any{"roles":items})
}

func (h ControlHandler) Permissions(w http.ResponseWriter, r *http.Request) {
	a,ok:=h.access(r)
	if !ok { writeJSON(w,http.StatusUnauthorized,map[string]string{"error":"unauthorized"}); return }
	if !a.Permissions["roles.manage"] { writeJSON(w,http.StatusForbidden,map[string]string{"error":"forbidden"}); return }
	ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second); defer cancel()
	rows,err:=h.DB.QueryContext(ctx,`SELECT key,description FROM control_permissions ORDER BY key`)
	if err!=nil { writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"}); return }
	defer rows.Close()
	type permission struct{ Key,Description string }
	items:=[]permission{}
	for rows.Next(){var x permission;if rows.Scan(&x.Key,&x.Description)==nil{items=append(items,x)}}
	writeJSON(w,http.StatusOK,map[string]any{"permissions":items})
}

func parsePGTextArray(raw string) []string {
	if raw=="{}" || len(raw)<2 { return []string{} }
	raw=raw[1:len(raw)-1]
	if raw=="" { return []string{} }
	return strings.Split(raw,",")
}


type controlSiteInput struct { Name string `json:"name"`; Slug string `json:"slug"`; Domain string `json:"domain"`; SiteType string `json:"siteType"`; ManagementMode string `json:"managementMode"`; Status string `json:"status"` }
type controlRoleInput struct { Name string `json:"name"`; Slug string `json:"slug"`; Description string `json:"description"`; Permissions []string `json:"permissions"` }
type controlScopeInput struct { AdminUserID string `json:"adminUserId"`; RoleID string `json:"roleId"`; OrganizationID string `json:"organizationId"`; SiteID *string `json:"siteId"` }

var slugRE=regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var domainRE=regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\\.)+[a-z]{2,63}$`)

func validSiteInput(in controlSiteInput) bool {
	types:=map[string]bool{"identity_platform":true,"local_directory":true,"web_archive":true,"service_site":true,"connected_platform":true}
	modes:=map[string]bool{"managed":true,"connected":true}
	statuses:=map[string]bool{"active":true,"disabled":true}
	return in.Name!="" && slugRE.MatchString(in.Slug) && domainRE.MatchString(in.Domain) && types[in.SiteType] && modes[in.ManagementMode] && statuses[in.Status]
}

func requestIP(r *http.Request) string {
	raw:=strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"),",")[0])
	if raw=="" { raw=r.RemoteAddr }
	if host,_,err:=net.SplitHostPort(raw); err==nil { raw=host }
	if net.ParseIP(raw)==nil { return "" }
	return raw
}

func auditTx(ctx context.Context,tx *sql.Tx,a controlAccess,r *http.Request,action,typ,id string,before,after any) error {
	b,_:=json.Marshal(before); n,_:=json.Marshal(after)
	var orgID string
	var siteID *string
	if typ=="site" { siteID=&id }
	_ = tx.QueryRowContext(ctx,`SELECT organization_id::text FROM control_admin_assignments WHERE admin_user_id=$1::uuid AND organization_id IS NOT NULL ORDER BY (site_id IS NULL) DESC LIMIT 1`,a.Admin.ID).Scan(&orgID)
	_,err:=tx.ExecContext(ctx,`INSERT INTO control_audit_logs(admin_user_id,organization_id,site_id,action,resource_type,resource_id,before_data,after_data,ip_address) VALUES($1::uuid,$2::uuid,$3::uuid,$4,$5,$6,$7::jsonb,$8::jsonb,NULLIF($9,'')::inet)`,a.Admin.ID,orgID,siteID,action,typ,id,string(b),string(n),requestIP(r))
	return err
}

func (h ControlHandler) CreateSite(w http.ResponseWriter,r *http.Request){a,ok:=h.access(r);if !ok{writeJSON(w,401,map[string]string{"error":"unauthorized"});return};if !a.Permissions["sites.manage"]{writeJSON(w,403,map[string]string{"error":"forbidden"});return};var in controlSiteInput;if json.NewDecoder(http.MaxBytesReader(w,r.Body,32<<10)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return};in.Name=strings.TrimSpace(in.Name);in.Slug=strings.ToLower(strings.TrimSpace(in.Slug));in.Domain=strings.ToLower(strings.TrimSpace(in.Domain));if !validSiteInput(in){writeJSON(w,400,map[string]string{"error":"invalid_site"});return};ctx,cancel:=context.WithTimeout(r.Context(),4*time.Second);defer cancel();tx,err:=h.DB.BeginTx(ctx,nil);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};defer tx.Rollback();var org,id string;err=tx.QueryRowContext(ctx,`SELECT organization_id::text FROM control_admin_assignments WHERE admin_user_id=$1::uuid AND organization_id IS NOT NULL AND site_id IS NULL LIMIT 1`,a.Admin.ID).Scan(&org);if err==nil{err=tx.QueryRowContext(ctx,`INSERT INTO sites(organization_id,slug,name,domain,site_type,management_mode,status) VALUES($1::uuid,$2,$3,$4,$5,$6,$7) RETURNING id::text`,org,in.Slug,in.Name,in.Domain,in.SiteType,in.ManagementMode,in.Status).Scan(&id)};if err==nil{err=auditTx(ctx,tx,a,r,"site.create","site",id,nil,in)};if err!=nil||tx.Commit()!=nil{writeJSON(w,400,map[string]string{"error":"site_create_failed"});return};writeJSON(w,201,map[string]any{"id":id})}

func (h ControlHandler) UpdateSite(w http.ResponseWriter,r *http.Request){a,ok:=h.access(r);if !ok{writeJSON(w,401,map[string]string{"error":"unauthorized"});return};if !a.Permissions["sites.manage"]{writeJSON(w,403,map[string]string{"error":"forbidden"});return};id:=r.PathValue("id");var in controlSiteInput;if json.NewDecoder(http.MaxBytesReader(w,r.Body,32<<10)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return};in.Name=strings.TrimSpace(in.Name);in.Slug=strings.ToLower(strings.TrimSpace(in.Slug));in.Domain=strings.ToLower(strings.TrimSpace(in.Domain));if !validSiteInput(in){writeJSON(w,400,map[string]string{"error":"invalid_site"});return};ctx,cancel:=context.WithTimeout(r.Context(),4*time.Second);defer cancel();tx,err:=h.DB.BeginTx(ctx,nil);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};defer tx.Rollback();var before map[string]any=map[string]any{};var oldName,oldDomain,oldType,oldMode,oldStatus string;err=tx.QueryRowContext(ctx,`SELECT name,domain,site_type,management_mode,status FROM sites WHERE id=$1::uuid AND id::text=ANY($2::text[])`,id,pqTextArray(a.SiteIDs)).Scan(&oldName,&oldDomain,&oldType,&oldMode,&oldStatus);before=map[string]any{"name":oldName,"domain":oldDomain,"siteType":oldType,"managementMode":oldMode,"status":oldStatus};if err==nil{_,err=tx.ExecContext(ctx,`UPDATE sites SET name=$1,domain=LOWER($2),site_type=$3,management_mode=$4,status=$5,updated_at=NOW() WHERE id=$6::uuid`,strings.TrimSpace(in.Name),strings.TrimSpace(in.Domain),in.SiteType,in.ManagementMode,in.Status,id)};if err==nil{err=auditTx(ctx,tx,a,r,"site.update","site",id,before,in)};if err!=nil||tx.Commit()!=nil{writeJSON(w,400,map[string]string{"error":"site_update_failed"});return};writeJSON(w,200,map[string]bool{"ok":true})}

func (h ControlHandler) CreateRole(w http.ResponseWriter,r *http.Request){a,ok:=h.access(r);if !ok{writeJSON(w,401,map[string]string{"error":"unauthorized"});return};if !a.Permissions["roles.manage"]{writeJSON(w,403,map[string]string{"error":"forbidden"});return};var in controlRoleInput;if json.NewDecoder(http.MaxBytesReader(w,r.Body,64<<10)).Decode(&in)!=nil||strings.TrimSpace(in.Name)==""||strings.TrimSpace(in.Slug)==""{writeJSON(w,400,map[string]string{"error":"invalid_request"});return};ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel();tx,err:=h.DB.BeginTx(ctx,nil);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};defer tx.Rollback();var org,id string;err=tx.QueryRowContext(ctx,`SELECT organization_id::text FROM control_admin_assignments WHERE admin_user_id=$1::uuid AND organization_id IS NOT NULL AND site_id IS NULL LIMIT 1`,a.Admin.ID).Scan(&org);if err==nil{err=tx.QueryRowContext(ctx,`INSERT INTO control_roles(organization_id,slug,name,description) VALUES($1::uuid,LOWER($2),$3,$4) RETURNING id::text`,org,strings.TrimSpace(in.Slug),strings.TrimSpace(in.Name),strings.TrimSpace(in.Description)).Scan(&id)};for _,p:=range in.Permissions{if err==nil{_,err=tx.ExecContext(ctx,`INSERT INTO control_role_permissions(role_id,permission_key) VALUES($1::uuid,$2)`,id,p)}};if err==nil{err=auditTx(ctx,tx,a,r,"role.create","role",id,nil,in)};if err!=nil||tx.Commit()!=nil{writeJSON(w,400,map[string]string{"error":"role_create_failed"});return};writeJSON(w,201,map[string]any{"id":id})}

func (h ControlHandler) UpdateRole(w http.ResponseWriter,r *http.Request){a,ok:=h.access(r);if !ok{writeJSON(w,401,map[string]string{"error":"unauthorized"});return};if !a.Permissions["roles.manage"]{writeJSON(w,403,map[string]string{"error":"forbidden"});return};id:=r.PathValue("id");var in controlRoleInput;if json.NewDecoder(http.MaxBytesReader(w,r.Body,64<<10)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return};ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel();tx,err:=h.DB.BeginTx(ctx,nil);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};defer tx.Rollback();var system bool;var oldName,oldDesc string;err=tx.QueryRowContext(ctx,`SELECT name,description,is_system FROM control_roles WHERE id=$1::uuid AND organization_id IN(SELECT organization_id FROM control_admin_assignments WHERE admin_user_id=$2::uuid)`,id,a.Admin.ID).Scan(&oldName,&oldDesc,&system);if system { writeJSON(w,http.StatusConflict,map[string]string{"error":"system_role_immutable"}); return };if err==nil{_,err=tx.ExecContext(ctx,`UPDATE control_roles SET name=$1,description=$2,updated_at=NOW() WHERE id=$3::uuid`,strings.TrimSpace(in.Name),strings.TrimSpace(in.Description),id)};if err==nil{_,err=tx.ExecContext(ctx,`DELETE FROM control_role_permissions WHERE role_id=$1::uuid`,id)};for _,p:=range in.Permissions{if err==nil{_,err=tx.ExecContext(ctx,`INSERT INTO control_role_permissions(role_id,permission_key) VALUES($1::uuid,$2)`,id,p)}};if err==nil{err=auditTx(ctx,tx,a,r,"role.update","role",id,map[string]any{"name":oldName,"description":oldDesc},in)};if err!=nil||tx.Commit()!=nil{writeJSON(w,400,map[string]string{"error":"role_update_failed"});return};writeJSON(w,200,map[string]bool{"ok":true})}

func (h ControlHandler) DeleteRole(w http.ResponseWriter,r *http.Request){a,ok:=h.access(r);if !ok{writeJSON(w,401,map[string]string{"error":"unauthorized"});return};if !a.Permissions["roles.manage"]{writeJSON(w,403,map[string]string{"error":"forbidden"});return};id:=r.PathValue("id");ctx,cancel:=context.WithTimeout(r.Context(),4*time.Second);defer cancel();tx,err:=h.DB.BeginTx(ctx,nil);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};defer tx.Rollback();var name string;var system bool;err=tx.QueryRowContext(ctx,`SELECT name,is_system FROM control_roles WHERE id=$1::uuid AND organization_id IN (SELECT organization_id FROM control_admin_assignments WHERE admin_user_id=$2::uuid AND organization_id IS NOT NULL)`,id,a.Admin.ID).Scan(&name,&system);if system{writeJSON(w,409,map[string]string{"error":"system_role"});return};if err==nil{_,err=tx.ExecContext(ctx,`DELETE FROM control_roles WHERE id=$1::uuid`,id)};if err==nil{err=auditTx(ctx,tx,a,r,"role.delete","role",id,map[string]any{"name":name},nil)};if err!=nil||tx.Commit()!=nil{writeJSON(w,400,map[string]string{"error":"role_delete_failed"});return};writeJSON(w,200,map[string]bool{"ok":true})}

func (h ControlHandler) AssignScope(w http.ResponseWriter,r *http.Request){a,ok:=h.access(r);if !ok{writeJSON(w,401,map[string]string{"error":"unauthorized"});return};if !a.Permissions["admins.manage"]&&!a.Permissions["roles.manage"]{writeJSON(w,403,map[string]string{"error":"forbidden"});return};var in controlScopeInput;if json.NewDecoder(http.MaxBytesReader(w,r.Body,32<<10)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return};ctx,cancel:=context.WithTimeout(r.Context(),4*time.Second);defer cancel();tx,err:=h.DB.BeginTx(ctx,nil);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};defer tx.Rollback();var allowedOrg bool;err=tx.QueryRowContext(ctx,`SELECT EXISTS(SELECT 1 FROM control_admin_assignments WHERE admin_user_id=$1::uuid AND organization_id=$2::uuid AND site_id IS NULL)`,a.Admin.ID,in.OrganizationID).Scan(&allowedOrg);if err==nil&&!allowedOrg{writeJSON(w,403,map[string]string{"error":"scope_escalation"});return};if err==nil&&in.SiteID!=nil{var allowedSite bool;err=tx.QueryRowContext(ctx,`SELECT EXISTS(SELECT 1 FROM sites WHERE id=$1::uuid AND organization_id=$2::uuid AND id::text=ANY($3::text[]))`,*in.SiteID,in.OrganizationID,pqTextArray(a.SiteIDs)).Scan(&allowedSite);if err==nil&&!allowedSite{writeJSON(w,403,map[string]string{"error":"scope_escalation"});return}};if err==nil{var roleSystem bool;err=tx.QueryRowContext(ctx,`SELECT is_system FROM control_roles WHERE id=$1::uuid AND organization_id=$2::uuid`,in.RoleID,in.OrganizationID).Scan(&roleSystem);if err==nil&&roleSystem&&a.Admin.Role!="super-admin"{writeJSON(w,403,map[string]string{"error":"system_role_assignment_forbidden"});return}};if in.AdminUserID==a.Admin.ID { var roleSlug string; if err==nil{err=tx.QueryRowContext(ctx,`SELECT slug FROM control_roles WHERE id=$1::uuid`,in.RoleID).Scan(&roleSlug)};if err==nil&&roleSlug!="super-admin"{writeJSON(w,409,map[string]string{"error":"self_privilege_change_forbidden"});return} };var id string;if err==nil{err=tx.QueryRowContext(ctx,`INSERT INTO control_admin_assignments(admin_user_id,role_id,organization_id,site_id) VALUES($1::uuid,$2::uuid,$3::uuid,$4::uuid) RETURNING id::text`,in.AdminUserID,in.RoleID,in.OrganizationID,in.SiteID).Scan(&id)};if err==nil{err=auditTx(ctx,tx,a,r,"scope.assign","admin_assignment",id,nil,in)};if err!=nil||tx.Commit()!=nil{writeJSON(w,400,map[string]string{"error":"scope_assign_failed"});return};writeJSON(w,201,map[string]any{"id":id})}

func (h ControlHandler) Admins(w http.ResponseWriter,r *http.Request){
	a,ok:=h.access(r);if !ok{writeJSON(w,401,map[string]string{"error":"unauthorized"});return}
	if !a.Permissions["admins.manage"]&&!a.Permissions["roles.manage"]{writeJSON(w,403,map[string]string{"error":"forbidden"});return}
	ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel()
	rows,err:=h.DB.QueryContext(ctx,`SELECT DISTINCT au.id::text,au.name,au.email,au.status FROM admin_users au JOIN control_admin_assignments ca ON ca.admin_user_id=au.id WHERE ca.organization_id IN (SELECT organization_id FROM control_admin_assignments WHERE admin_user_id=$1::uuid AND organization_id IS NOT NULL) ORDER BY au.email`,a.Admin.ID)
	if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};defer rows.Close()
	type item struct{ID,Name,Email,Status string};items:=[]item{};for rows.Next(){var x item;if rows.Scan(&x.ID,&x.Name,&x.Email,&x.Status)==nil{items=append(items,x)}};writeJSON(w,200,map[string]any{"admins":items})
}

func (h ControlHandler) Assignments(w http.ResponseWriter,r *http.Request){
	a,ok:=h.access(r);if !ok{writeJSON(w,401,map[string]string{"error":"unauthorized"});return}
	if !a.Permissions["admins.manage"]&&!a.Permissions["roles.manage"]{writeJSON(w,403,map[string]string{"error":"forbidden"});return}
	ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel()
	rows,err:=h.DB.QueryContext(ctx,`SELECT ca.id::text,au.id::text,au.name,au.email,r.id::text,r.name,r.slug,ca.organization_id::text,COALESCE(ca.site_id::text,''),COALESCE(s.name,'') FROM control_admin_assignments ca JOIN admin_users au ON au.id=ca.admin_user_id JOIN control_roles r ON r.id=ca.role_id LEFT JOIN sites s ON s.id=ca.site_id WHERE ca.organization_id IN (SELECT organization_id FROM control_admin_assignments WHERE admin_user_id=$1::uuid AND organization_id IS NOT NULL) ORDER BY au.email,r.name,s.name NULLS FIRST`,a.Admin.ID)
	if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};defer rows.Close()
	type item struct{ID,AdminUserID,AdminName,AdminEmail,RoleID,RoleName,RoleSlug,OrganizationID,SiteID,SiteName string}
	items:=[]item{};for rows.Next(){var x item;if rows.Scan(&x.ID,&x.AdminUserID,&x.AdminName,&x.AdminEmail,&x.RoleID,&x.RoleName,&x.RoleSlug,&x.OrganizationID,&x.SiteID,&x.SiteName)==nil{items=append(items,x)}}
	writeJSON(w,200,map[string]any{"assignments":items})
}

func (h ControlHandler) RevokeScope(w http.ResponseWriter,r *http.Request){
	a,ok:=h.access(r); if !ok { writeJSON(w,401,map[string]string{"error":"unauthorized"}); return }
	if !a.Permissions["admins.manage"]&&!a.Permissions["roles.manage"] { writeJSON(w,403,map[string]string{"error":"forbidden"}); return }
	id:=r.PathValue("id")
	ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second); defer cancel()
	tx,err:=h.DB.BeginTx(ctx,&sql.TxOptions{Isolation:sql.LevelSerializable}); if err!=nil { writeJSON(w,500,map[string]string{"error":"internal_error"}); return }; defer tx.Rollback()
	var targetAdmin,roleID,orgID string; var siteID sql.NullString; var roleSlug string
	err=tx.QueryRowContext(ctx,`SELECT ca.admin_user_id::text,ca.role_id::text,ca.organization_id::text,ca.site_id::text,r.slug FROM control_admin_assignments ca JOIN control_roles r ON r.id=ca.role_id WHERE ca.id=$1::uuid AND ca.organization_id IN (SELECT organization_id FROM control_admin_assignments WHERE admin_user_id=$2::uuid AND organization_id IS NOT NULL)`,id,a.Admin.ID).Scan(&targetAdmin,&roleID,&orgID,&siteID,&roleSlug)
	if err!=nil { writeJSON(w,404,map[string]string{"error":"assignment_not_found"}); return }
	if targetAdmin==a.Admin.ID { writeJSON(w,409,map[string]string{"error":"self_scope_revoke_forbidden"}); return }
	if roleSlug=="super-admin" && !siteID.Valid {
		var remaining int
		err=tx.QueryRowContext(ctx,`SELECT count(DISTINCT ca.admin_user_id) FROM control_admin_assignments ca JOIN control_roles r ON r.id=ca.role_id JOIN admin_users au ON au.id=ca.admin_user_id WHERE ca.organization_id=$1::uuid AND ca.site_id IS NULL AND r.slug='super-admin' AND au.status='active' AND ca.id<>$2::uuid`,orgID,id).Scan(&remaining)
		if err!=nil { writeJSON(w,500,map[string]string{"error":"internal_error"}); return }
		if remaining<1 { writeJSON(w,409,map[string]string{"error":"last_super_admin"}); return }
	}
	before:=map[string]any{"adminUserId":targetAdmin,"roleId":roleID,"organizationId":orgID}
	if siteID.Valid { before["siteId"]=siteID.String }
	res,err:=tx.ExecContext(ctx,`DELETE FROM control_admin_assignments WHERE id=$1::uuid`,id); if err!=nil { writeJSON(w,400,map[string]string{"error":"scope_revoke_failed"}); return }
	n,_:=res.RowsAffected(); if n!=1 { writeJSON(w,404,map[string]string{"error":"assignment_not_found"}); return }
	if err=auditTx(ctx,tx,a,r,"scope.revoke","admin_assignment",id,before,nil);err!=nil { writeJSON(w,500,map[string]string{"error":"audit_failed"}); return }
	if err=tx.Commit();err!=nil { writeJSON(w,409,map[string]string{"error":"scope_revoke_conflict"}); return }
	writeJSON(w,200,map[string]bool{"ok":true})
}

func (h ControlHandler) Audit(w http.ResponseWriter,r *http.Request){a,ok:=h.access(r);if !ok{writeJSON(w,401,map[string]string{"error":"unauthorized"});return};if !a.Permissions["audit.view"]{writeJSON(w,403,map[string]string{"error":"forbidden"});return};ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel();rows,err:=h.DB.QueryContext(ctx,`SELECT l.id,l.action,l.resource_type,COALESCE(l.resource_id,''),COALESCE(a.email,''),l.created_at FROM control_audit_logs l LEFT JOIN admin_users a ON a.id=l.admin_user_id WHERE l.admin_user_id=$1::uuid OR l.organization_id IN (SELECT organization_id FROM control_admin_assignments WHERE admin_user_id=$1::uuid AND organization_id IS NOT NULL) ORDER BY l.created_at DESC LIMIT 100`,a.Admin.ID);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};defer rows.Close();type item struct{ID int64;Action,ResourceType,ResourceID,Actor string;CreatedAt time.Time};items:=[]item{};for rows.Next(){var x item;if rows.Scan(&x.ID,&x.Action,&x.ResourceType,&x.ResourceID,&x.Actor,&x.CreatedAt)==nil{items=append(items,x)}};writeJSON(w,200,map[string]any{"audit":items})}
