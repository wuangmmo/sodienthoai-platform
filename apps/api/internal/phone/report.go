package phone

import (
 "context"
 "errors"
 "strings"
)

var ErrInvalidReport = errors.New("invalid phone report")
var allowedReasons = map[string]bool{"spam":true,"scam":true,"telemarketing":true,"harassment":true,"wrong_identity":true,"other":true}

type ReportInput struct { Reason string `json:"reason"`; Comment string `json:"comment"` }

func (s Service) Report(ctx context.Context, e164 string, in ReportInput, reporterHash string) error {
 in.Reason=strings.TrimSpace(strings.ToLower(in.Reason)); in.Comment=strings.TrimSpace(in.Comment)
 if !allowedReasons[in.Reason] || len(in.Comment)>2000 { return ErrInvalidReport }
 n,err:=s.Repository.FindByE164(ctx,e164); if err!=nil{return err}
 return s.Repository.CreateReport(ctx,n.ID,e164,in,reporterHash)
}

func (r Repository) CreateReport(ctx context.Context, phoneID,e164 string,in ReportInput,reporterHash string) error {
 _,err:=r.DB.ExecContext(ctx,`INSERT INTO phone_reports(phone_number_id,e164,reason,comment,reporter_hash) VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,''))`,phoneID,e164,in.Reason,in.Comment,reporterHash)
 return err
}
