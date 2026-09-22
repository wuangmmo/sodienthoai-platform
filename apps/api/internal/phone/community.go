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


type PublicComment struct{ID string `json:"id"`;Body string `json:"body"`;HelpfulCount int `json:"helpful_count"`;CreatedAt string `json:"created_at"`}
func(s Service) PublicComments(ctx context.Context,e164 string,limit int)([]PublicComment,error){
 n,err:=s.Repository.FindByE164(ctx,e164);if err!=nil{return nil,err};if limit<1||limit>100{limit=30};rows,err:=s.Repository.DB.QueryContext(ctx,`SELECT id::text,body,helpful_count,created_at::text FROM phone_comments WHERE phone_number_id=$1 AND status='approved' ORDER BY helpful_count DESC,created_at DESC LIMIT $2`,n.ID,limit);if err!=nil{return nil,err};defer rows.Close();out:=[]PublicComment{};for rows.Next(){var x PublicComment;if err:=rows.Scan(&x.ID,&x.Body,&x.HelpfulCount,&x.CreatedAt);err!=nil{return nil,err};out=append(out,x)};return out,rows.Err()
}
func(s Service) HelpfulComment(ctx context.Context,subject,commentID string)error{
 uid,err:=s.EnsureUser(ctx,subject);if err!=nil{return err};res,err:=s.Repository.DB.ExecContext(ctx,`INSERT INTO comment_helpful_votes(comment_id,user_id) VALUES($1,$2) ON CONFLICT DO NOTHING`,commentID,uid);if err!=nil{return err};if n,_:=res.RowsAffected();n>0{_,err=s.Repository.DB.ExecContext(ctx,`UPDATE phone_comments SET helpful_count=helpful_count+1 WHERE id=$1 AND status='approved'`,commentID)};return err
}
