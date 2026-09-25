package phone

import (
 "context"
 "errors"
 "strings"
)

var ErrInvalidBusinessModeration=errors.New("invalid business moderation")

func(s Service) ModerateAppeal(ctx context.Context,id,status,actor,note string)error{
 status=strings.ToLower(strings.TrimSpace(status));if status!="accepted"&&status!="rejected"{return ErrInvalidBusinessModeration}
 res,err:=s.Repository.DB.ExecContext(ctx,`UPDATE phone_appeals SET status=$2,reviewed_at=NOW(),reviewed_by=$3,resolution_note=NULLIF($4,'') WHERE id=$1 AND status='pending'`,id,status,actor,strings.TrimSpace(note));if err!=nil{return err};n,_:=res.RowsAffected();if n==0{return ErrInvalidBusinessModeration}
 return s.Repository.AuditAdmin(ctx,actor,"moderate_appeal","phone_appeal",id,map[string]any{"status":status})
}

func(s Service) ModerateBusinessVerification(ctx context.Context,id,status,actor,note string)error{
 status=strings.ToLower(strings.TrimSpace(status));if status!="approved"&&status!="rejected"{return ErrInvalidBusinessModeration}
 tx,err:=s.Repository.DB.BeginTx(ctx,nil);if err!=nil{return err};defer tx.Rollback()
 var businessID,userID string
 err=tx.QueryRowContext(ctx,`UPDATE business_verification_requests SET status=$2,reviewed_at=NOW(),reviewed_by=$3,resolution_note=NULLIF($4,'') WHERE id=$1 AND status='pending' RETURNING business_id::text,user_id::text`,id,status,actor,strings.TrimSpace(note)).Scan(&businessID,&userID);if err!=nil{return ErrInvalidBusinessModeration}
 if status=="approved" {
  res,err:=tx.ExecContext(ctx,`UPDATE businesses SET verification_status='verified',updated_at=NOW() WHERE id=$1 AND verification_status='pending'`,businessID);if err!=nil{return err};n,_:=res.RowsAffected();if n!=1{return ErrInvalidBusinessModeration}
  res,err=tx.ExecContext(ctx,`UPDATE business_ownerships SET status='verified',verified_at=COALESCE(verified_at,NOW()) WHERE business_id=$1 AND user_id=$2 AND status IN ('pending','verified')`,businessID,userID);if err!=nil{return err};n,_=res.RowsAffected();if n!=1{return ErrInvalidBusinessModeration}

 } else {
  res,err:=tx.ExecContext(ctx,`UPDATE businesses SET verification_status='rejected',updated_at=NOW() WHERE id=$1 AND verification_status='pending'`,businessID);if err!=nil{return err};n,_:=res.RowsAffected();if n!=1{return ErrInvalidBusinessModeration}
  res,err=tx.ExecContext(ctx,`UPDATE business_ownerships SET status='rejected' WHERE business_id=$1 AND user_id=$2 AND status IN ('pending','verified')`,businessID,userID);if err!=nil{return err};n,_=res.RowsAffected();if n!=1{return ErrInvalidBusinessModeration}
 }
 if err=tx.Commit();err!=nil{return err}
 return s.Repository.AuditAdmin(ctx,actor,"moderate_business_verification","business_verification",id,map[string]any{"status":status,"business_id":businessID})
}

func(s Service) ModerateBusinessReview(ctx context.Context,id,status,actor string)error{
 status=strings.ToLower(strings.TrimSpace(status));if status!="approved"&&status!="rejected"&&status!="removed"{return ErrInvalidBusinessModeration}
 res,err:=s.Repository.DB.ExecContext(ctx,`UPDATE business_reviews SET status=$2,reviewed_at=NOW(),reviewed_by=$3,updated_at=NOW() WHERE id=$1 AND ((status='pending' AND $2 IN ('approved','rejected')) OR (status='approved' AND $2='removed'))`,id,status,actor);if err!=nil{return err};n,_:=res.RowsAffected();if n==0{return ErrInvalidBusinessModeration}
 return s.Repository.AuditAdmin(ctx,actor,"moderate_business_review","business_review",id,map[string]any{"status":status})
}
