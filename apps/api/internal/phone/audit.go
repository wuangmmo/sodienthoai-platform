package phone
import("context";"encoding/json")
func (r Repository)AuditAdmin(ctx context.Context,actor,action,resourceType,resourceID string,metadata any)error{raw,err:=json.Marshal(metadata);if err!=nil{return err};_,err=r.DB.ExecContext(ctx,`INSERT INTO admin_audit_log(actor,action,resource_type,resource_id,metadata) VALUES($1,$2,$3,NULLIF($4,''),$5::jsonb)`,actor,action,resourceType,resourceID,string(raw));return err}
