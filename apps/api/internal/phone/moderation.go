package phone

import ("context";"errors";"strings")
var ErrInvalidModeration=errors.New("invalid moderation")
func (s Service) ModerateReport(ctx context.Context,id,status string) error {
 status=strings.ToLower(strings.TrimSpace(status));if status!="approved"&&status!="rejected"{return ErrInvalidModeration}
 var phoneID string
 err:=s.Repository.DB.QueryRowContext(ctx,"UPDATE phone_reports SET status=$2,reviewed_at=NOW() WHERE id=$1 AND status='pending' RETURNING phone_number_id::text",id,status).Scan(&phoneID);if err!=nil{return err}
 if _,err=s.Repository.DB.ExecContext(ctx,"SELECT refresh_phone_reputation($1)",phoneID);err!=nil{return err}
 _,_=s.Repository.DB.ExecContext(ctx,"INSERT INTO phone_reputation_events(phone_number_id,event_type,source) VALUES($1,$2,$3)",phoneID,"report_"+status,"moderation")
 _ = s.Repository.AuditAdmin(ctx,"admin-api","moderate_report","phone_report",id,map[string]string{"status":status,"phone_id":phoneID})
 if err:=s.Repository.RefreshDerivedStatus(ctx,phoneID);err!=nil{return err};s.invalidatePhoneCache(ctx,phoneID);return nil
}
func (s Service) ModerateClaim(ctx context.Context,id,status string) error {
 status=strings.ToLower(strings.TrimSpace(status));if status!="verified"&&status!="rejected"{return ErrInvalidModeration}
 var phoneID string
 err:=s.Repository.DB.QueryRowContext(ctx,"UPDATE phone_claims SET status=$2,reviewed_at=NOW() WHERE id=$1 AND status='pending' RETURNING phone_number_id::text",id,status).Scan(&phoneID);if err!=nil{return err}
 if status=="verified"{_,err=s.Repository.DB.ExecContext(ctx,"UPDATE phone_numbers SET verification_status='verified',updated_at=NOW() WHERE id=$1",phoneID);if err!=nil{return err}}
 _,_=s.Repository.DB.ExecContext(ctx,"INSERT INTO phone_reputation_events(phone_number_id,event_type,source) VALUES($1,$2,$3)",phoneID,"claim_"+status,"moderation")
 _ = s.Repository.AuditAdmin(ctx,"admin-api","moderate_claim","phone_claim",id,map[string]string{"status":status,"phone_id":phoneID})
 if err:=s.Repository.RefreshDerivedStatus(ctx,phoneID);err!=nil{return err};s.invalidatePhoneCache(ctx,phoneID);return nil
}
