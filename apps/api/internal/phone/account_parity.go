package phone

import (
 "context"
 "errors"
 "strings"
)

var ErrInvalidPreference = errors.New("invalid phone preference")
var ErrInvalidAppeal = errors.New("invalid phone appeal")
var ErrDuplicateAppeal = errors.New("pending phone appeal already exists")

type UserPhonePreference struct {
 E164 string `json:"e164"`
 Disposition string `json:"disposition"`
 UpdatedAt string `json:"updated_at"`
}

type UserLookup struct {
 ID string `json:"id"`
 E164 string `json:"e164"`
 LookedUpAt string `json:"looked_up_at"`
}

type AppealInput struct {
 Reason string `json:"reason"`
 Statement string `json:"statement"`
}

func (s Service) SetPhonePreference(ctx context.Context, subject, e164, disposition string) error {
 disposition = strings.ToLower(strings.TrimSpace(disposition))
 if disposition != "trusted" && disposition != "blocked" { return ErrInvalidPreference }
 uid,err:=s.EnsureUser(ctx,subject);if err!=nil{return err}
 n,err:=s.Repository.FindByE164(ctx,e164);if err!=nil{return err}
 _,err=s.Repository.DB.ExecContext(ctx,`INSERT INTO user_phone_preferences(user_id,phone_number_id,disposition) VALUES($1,$2,$3)
 ON CONFLICT(user_id,phone_number_id) DO UPDATE SET disposition=EXCLUDED.disposition,updated_at=NOW()`,uid,n.ID,disposition)
 return err
}

func (s Service) DeletePhonePreference(ctx context.Context, subject, e164 string) error {
 uid,err:=s.EnsureUser(ctx,subject);if err!=nil{return err}
 n,err:=s.Repository.FindByE164(ctx,e164);if err!=nil{return err}
 _,err=s.Repository.DB.ExecContext(ctx,`DELETE FROM user_phone_preferences WHERE user_id=$1 AND phone_number_id=$2`,uid,n.ID)
 return err
}

func (s Service) PhonePreferences(ctx context.Context, subject, disposition string) ([]UserPhonePreference,error) {
 uid,err:=s.EnsureUser(ctx,subject);if err!=nil{return nil,err}
 args:=[]any{uid};q:=`SELECT n.e164,p.disposition,p.updated_at::text FROM user_phone_preferences p JOIN phone_numbers n ON n.id=p.phone_number_id WHERE p.user_id=$1`
 disposition=strings.ToLower(strings.TrimSpace(disposition))
 if disposition!="" { if disposition!="trusted"&&disposition!="blocked"{return nil,ErrInvalidPreference};q+=" AND p.disposition=$2";args=append(args,disposition) }
 q+=" ORDER BY p.updated_at DESC LIMIT 500"
 rows,err:=s.Repository.DB.QueryContext(ctx,q,args...);if err!=nil{return nil,err};defer rows.Close()
 out:=[]UserPhonePreference{};for rows.Next(){var x UserPhonePreference;if err:=rows.Scan(&x.E164,&x.Disposition,&x.UpdatedAt);err!=nil{return nil,err};out=append(out,x)}
 return out,rows.Err()
}

func (s Service) RecordUserLookup(ctx context.Context, subject, e164 string) error {
 uid,err:=s.EnsureUser(ctx,subject);if err!=nil{return err}
 var phoneID any
 n,findErr:=s.Repository.FindByE164(ctx,e164);if findErr==nil{phoneID=n.ID}else if !IsNotFound(findErr){return findErr}
 _,err=s.Repository.DB.ExecContext(ctx,`INSERT INTO user_lookup_history(user_id,phone_number_id,e164) VALUES($1,$2,$3)`,uid,phoneID,e164)
 return err
}

func (s Service) UserLookupHistory(ctx context.Context, subject string) ([]UserLookup,error) {
 uid,err:=s.EnsureUser(ctx,subject);if err!=nil{return nil,err}
 rows,err:=s.Repository.DB.QueryContext(ctx,`SELECT id::text,e164,looked_up_at::text FROM user_lookup_history WHERE user_id=$1 ORDER BY looked_up_at DESC LIMIT 200`,uid);if err!=nil{return nil,err};defer rows.Close()
 out:=[]UserLookup{};for rows.Next(){var x UserLookup;if err:=rows.Scan(&x.ID,&x.E164,&x.LookedUpAt);err!=nil{return nil,err};out=append(out,x)}
 return out,rows.Err()
}

func (s Service) DeleteUserLookupHistory(ctx context.Context, subject string) error {
 uid,err:=s.EnsureUser(ctx,subject);if err!=nil{return err}
 _,err=s.Repository.DB.ExecContext(ctx,`DELETE FROM user_lookup_history WHERE user_id=$1`,uid);return err
}

func (s Service) CreateAppeal(ctx context.Context, subject, e164 string, in AppealInput) (string,error) {
 in.Reason=strings.ToLower(strings.TrimSpace(in.Reason));in.Statement=strings.TrimSpace(in.Statement)
 allowed:=map[string]bool{"wrong_identity":true,"wrong_label":true,"resolved":true,"privacy":true,"other":true}
 if !allowed[in.Reason]||len(in.Statement)<10||len(in.Statement)>4000{return "",ErrInvalidAppeal}
 uid,err:=s.EnsureUser(ctx,subject);if err!=nil{return "",err}
 n,err:=s.Repository.FindByE164(ctx,e164);if err!=nil{return "",err}
 var id string
 err=s.Repository.DB.QueryRowContext(ctx,`INSERT INTO phone_appeals(phone_number_id,user_id,reason,statement) VALUES($1,$2,$3,$4) RETURNING id::text`,n.ID,uid,in.Reason,in.Statement).Scan(&id)
 if err!=nil && strings.Contains(strings.ToLower(err.Error()),"idx_phone_appeals_one_pending"){return "",ErrDuplicateAppeal}
 return id,err
}


type UserAppeal struct {
 ID string `json:"id"`
 E164 string `json:"e164"`
 Reason string `json:"reason"`
 Statement string `json:"statement"`
 Status string `json:"status"`
 CreatedAt string `json:"created_at"`
}

func (s Service) UserAppeals(ctx context.Context, subject string) ([]UserAppeal,error) {
 uid,err:=s.EnsureUser(ctx,subject);if err!=nil{return nil,err}
 rows,err:=s.Repository.DB.QueryContext(ctx,`SELECT a.id::text,n.e164,a.reason,a.statement,a.status,a.created_at::text FROM phone_appeals a JOIN phone_numbers n ON n.id=a.phone_number_id WHERE a.user_id=$1 ORDER BY a.created_at DESC LIMIT 200`,uid);if err!=nil{return nil,err};defer rows.Close()
 out:=[]UserAppeal{};for rows.Next(){var x UserAppeal;if err:=rows.Scan(&x.ID,&x.E164,&x.Reason,&x.Statement,&x.Status,&x.CreatedAt);err!=nil{return nil,err};out=append(out,x)}
 return out,rows.Err()
}

func (s Service) WithdrawAppeal(ctx context.Context, subject, id string) error {
 uid,err:=s.EnsureUser(ctx,subject);if err!=nil{return err}
 res,err:=s.Repository.DB.ExecContext(ctx,`UPDATE phone_appeals SET status='cancelled' WHERE id=$1 AND user_id=$2 AND status='pending'`,id,uid);if err!=nil{return err}
 n,_:=res.RowsAffected();if n==0{return ErrInvalidAppeal};return nil
}
