package httpserver

import (
	"context"
	"database/sql"
	"net/http"
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
