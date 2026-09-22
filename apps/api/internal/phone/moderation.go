package phone

import ("context";"errors";"strings")
var ErrInvalidModeration=errors.New("invalid moderation")
func (s Service) ModerateReport(ctx context.Context,id,status string) error {
 status=strings.ToLower(strings.TrimSpace(status));if status!="approved"&&status!="rejected"{return ErrInvalidModeration}
 var phoneID string
 err:=s.Repository.DB.QueryRowContext(ctx,"UPDATE phone_reports SET status=$2,reviewed_at=NOW() WHERE id=$1 RETURNING phone_number_id::text",id,status).Scan(&phoneID);if err!=nil{return err}
 if status=="approved"{if _,err=s.Repository.DB.ExecContext(ctx,"SELECT refresh_phone_reputation($1)",phoneID);err!=nil{return err}}
 return s.Repository.RefreshDerivedStatus(ctx,phoneID)
}
func (s Service) ModerateClaim(ctx context.Context,id,status string) error {
 status=strings.ToLower(strings.TrimSpace(status));if status!="verified"&&status!="rejected"{return ErrInvalidModeration}
 var phoneID string
 err:=s.Repository.DB.QueryRowContext(ctx,"UPDATE phone_claims SET status=$2,reviewed_at=NOW() WHERE id=$1 RETURNING phone_number_id::text",id,status).Scan(&phoneID);if err!=nil{return err}
 if status=="verified"{_,err=s.Repository.DB.ExecContext(ctx,"UPDATE phone_numbers SET verification_status='verified',updated_at=NOW() WHERE id=$1",phoneID);if err!=nil{return err}}
 return s.Repository.RefreshDerivedStatus(ctx,phoneID)
}
