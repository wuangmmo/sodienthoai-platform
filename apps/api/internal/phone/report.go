package phone

import (
 "context"
 "database/sql"
 "errors"
 "strings"
)

var ErrInvalidReport = errors.New("invalid phone report")
var ErrDuplicateReport = errors.New("duplicate phone report")
var allowedReasons = map[string]bool{"spam":true,"scam":true,"telemarketing":true,"harassment":true,"impersonation":true,"debt_collection":true,"wrong_identity":true,"other":true}

type ReportInput struct { Reason string `json:"reason"`; Comment string `json:"comment"` }

func (s Service) Report(ctx context.Context, e164 string, in ReportInput, reporterHash string) error {
 in.Reason=strings.TrimSpace(strings.ToLower(in.Reason)); in.Comment=strings.TrimSpace(in.Comment)
 if !allowedReasons[in.Reason] || len(in.Comment)>2000 { return ErrInvalidReport }
 n,err:=s.Repository.FindByE164(ctx,e164); if err!=nil{return err}
 return s.Repository.CreateReport(ctx,n.ID,e164,in,reporterHash)
}

func (r Repository) CreateReport(ctx context.Context, phoneID,e164 string,in ReportInput,reporterHash string) error {
 tx,err:=r.DB.BeginTx(ctx,nil);if err!=nil{return err};defer tx.Rollback()
 if reporterHash!="" {
  var ok bool
  err=tx.QueryRowContext(ctx,`INSERT INTO phone_report_fingerprints(phone_number_id,reporter_hash,reason,last_reported_at)
VALUES($1,$2,$3,NOW())
ON CONFLICT(phone_number_id,reporter_hash,reason) DO UPDATE SET last_reported_at=EXCLUDED.last_reported_at
WHERE phone_report_fingerprints.last_reported_at < NOW()-INTERVAL '24 hours'
RETURNING TRUE`,phoneID,reporterHash,in.Reason).Scan(&ok)
  if errors.Is(err,sql.ErrNoRows){return ErrDuplicateReport};if err!=nil{return err}
 }
 _,err=tx.ExecContext(ctx,`INSERT INTO phone_reports(phone_number_id,e164,reason,comment,reporter_hash) VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,''))`,phoneID,e164,in.Reason,in.Comment,reporterHash)
 if err!=nil{return err};return tx.Commit()
}
