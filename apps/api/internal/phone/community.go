package phone

import (
 "context"
 "errors"
 "strings"
)

var ErrInvalidComment=errors.New("invalid comment")

func(s Service) EnsureUser(ctx context.Context,subject string)(string,error){
 subject=strings.TrimSpace(subject);if subject==""||len(subject)>128{return "",ErrInvalidComment};var id string
 err:=s.Repository.DB.QueryRowContext(ctx,`INSERT INTO user_accounts(subject_key) VALUES($1) ON CONFLICT(subject_key) DO UPDATE SET updated_at=NOW() RETURNING id::text`,subject).Scan(&id);return id,err
}
func(s Service) Follow(ctx context.Context,subject,e164 string,follow bool)error{
 uid,err:=s.EnsureUser(ctx,subject);if err!=nil{return err};n,err:=s.Repository.FindByE164(ctx,e164);if err!=nil{return err}
 if follow{_,err=s.Repository.DB.ExecContext(ctx,`INSERT INTO phone_follows(user_id,phone_number_id) VALUES($1,$2) ON CONFLICT DO NOTHING`,uid,n.ID)}else{_,err=s.Repository.DB.ExecContext(ctx,`DELETE FROM phone_follows WHERE user_id=$1 AND phone_number_id=$2`,uid,n.ID)};return err
}
func(s Service) Comment(ctx context.Context,subject,e164,body string)(string,error){
 body=strings.TrimSpace(body);if len(body)<1||len(body)>1000{return "",ErrInvalidComment};uid,err:=s.EnsureUser(ctx,subject);if err!=nil{return "",err};n,err:=s.Repository.FindByE164(ctx,e164);if err!=nil{return "",err};var id string
 err=s.Repository.DB.QueryRowContext(ctx,`INSERT INTO phone_comments(phone_number_id,user_id,body) VALUES($1,$2,$3) RETURNING id::text`,n.ID,uid,body).Scan(&id);return id,err
}
