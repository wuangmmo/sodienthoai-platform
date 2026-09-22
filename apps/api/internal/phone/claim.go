package phone

import (
 "context"
 "errors"
 "strings"
)

var ErrInvalidClaim = errors.New("invalid phone claim")
type ClaimInput struct { Name string `json:"name"`; Kind string `json:"kind"`; Email string `json:"email"`; Evidence string `json:"evidence"` }

func (s Service) Claim(ctx context.Context,e164 string,in ClaimInput) error {
 in.Name=strings.TrimSpace(in.Name);in.Kind=strings.TrimSpace(strings.ToLower(in.Kind));in.Email=strings.TrimSpace(in.Email);in.Evidence=strings.TrimSpace(in.Evidence)
 if in.Name==""||len(in.Name)>255||len(in.Email)>320||len(in.Evidence)>4000{return ErrInvalidClaim}
 if in.Kind==""{in.Kind="unknown"}; if in.Kind!="person"&&in.Kind!="business"&&in.Kind!="organization"&&in.Kind!="service"&&in.Kind!="unknown"{return ErrInvalidClaim}
 n,err:=s.Repository.FindByE164(ctx,e164);if err!=nil{return err}
 _,err=s.Repository.DB.ExecContext(ctx,`INSERT INTO phone_claims(phone_number_id,claimant_name,claimant_type,contact_email,evidence_note) VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,''))`,n.ID,in.Name,in.Kind,in.Email,in.Evidence)
 return err
}
