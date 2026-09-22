package phone

import (
 "context"
 "strings"
)

type Contact struct {
 ID string `json:"id"`
 LocalName *string `json:"local_name,omitempty"`
 RawPhone string `json:"raw_phone"`
 E164 *string `json:"e164,omitempty"`
 PhoneID *string `json:"phone_id,omitempty"`
 MatchStatus string `json:"match_status"`
 ReputationLabel *string `json:"reputation_label,omitempty"`
 PublicIdentity *Identity `json:"public_identity,omitempty"`
}

type ContactImportItem struct { Name string `json:"name"`; Phone string `json:"phone"` }
type ContactImportResult struct { Imported int `json:"imported"`; Skipped int `json:"skipped"`; Contacts []Contact `json:"contacts"` }

func contactMatch(localName string, identity *Identity, disputed bool) string {
 if disputed{return "disputed"};if identity==nil{return "unknown"}
 a:=strings.ToLower(strings.TrimSpace(localName));b:=strings.ToLower(strings.TrimSpace(identity.DisplayName))
 if a==""{return "possible_match"};if a==b{return "match"}
 if strings.Contains(a,b)||strings.Contains(b,a){return "possible_match"};return "mismatch"
}

func (s Service) ImportContacts(ctx context.Context,ownerKey string,items []ContactImportItem)(ContactImportResult,error){
 ownerKey=strings.TrimSpace(ownerKey);if ownerKey==""||len(items)==0||len(items)>5000{return ContactImportResult{},ErrInvalidClaim}
 var bookID string
 err:=s.Repository.DB.QueryRowContext(ctx,`INSERT INTO contact_books(owner_key) VALUES($1) RETURNING id::text`,ownerKey).Scan(&bookID);if err!=nil{return ContactImportResult{},err}
 out:=ContactImportResult{Contacts:[]Contact{}}
 for _,item:=range items{
  raw:=strings.TrimSpace(item.Phone);name:=strings.TrimSpace(item.Name);if raw==""{out.Skipped++;continue}
  e164,nerr:=NormalizeForCountry(raw,"VN");if nerr!=nil{out.Skipped++;continue}
  n,err:=s.Repository.FindByE164(ctx,e164);var phoneID *string;status:="unknown";var identity *Identity;var rep *string
  if err==nil{phoneID=&n.ID;rep=&n.ReputationLabel;ids,_:=s.Repository.IdentitiesByPhoneID(ctx,n.ID);disputed,_:=s.Repository.HasOpenIdentityDispute(ctx,n.ID);for i:=range ids{if ids[i].IsPrimary{identity=&ids[i];break}};status=contactMatch(name,identity,disputed)}
  var id string;err=s.Repository.DB.QueryRowContext(ctx,`INSERT INTO private_contacts(contact_book_id,local_name,raw_phone,normalized_e164,phone_number_id,match_status,last_checked_at) VALUES($1,NULLIF($2,''),$3,$4,$5,$6,NOW()) ON CONFLICT(contact_book_id,normalized_e164) DO UPDATE SET local_name=EXCLUDED.local_name,raw_phone=EXCLUDED.raw_phone,phone_number_id=EXCLUDED.phone_number_id,match_status=EXCLUDED.match_status,last_checked_at=NOW(),updated_at=NOW() RETURNING id::text`,bookID,name,raw,e164,phoneID,status).Scan(&id);if err!=nil{return out,err}
  e:=e164;out.Contacts=append(out.Contacts,Contact{ID:id,LocalName:nil,RawPhone:raw,E164:&e,PhoneID:phoneID,MatchStatus:status,ReputationLabel:rep,PublicIdentity:identity});out.Imported++
 }
 return out,nil
}


func(s Service) ListContacts(ctx context.Context,ownerKey,status string,limit int)([]Contact,error){
 if limit<1||limit>500{limit=100};status=strings.ToLower(strings.TrimSpace(status))
 rows,err:=s.Repository.DB.QueryContext(ctx,`SELECT c.id::text,c.local_name,c.raw_phone,c.normalized_e164,c.phone_number_id::text,c.match_status::text,n.reputation_label
FROM private_contacts c JOIN contact_books b ON b.id=c.contact_book_id LEFT JOIN phone_numbers n ON n.id=c.phone_number_id
WHERE b.owner_key=$1 AND ($2='' OR c.match_status::text=$2) ORDER BY c.updated_at DESC LIMIT $3`,ownerKey,status,limit);if err!=nil{return nil,err};defer rows.Close()
 out:=[]Contact{};for rows.Next(){var x Contact;if err:=rows.Scan(&x.ID,&x.LocalName,&x.RawPhone,&x.E164,&x.PhoneID,&x.MatchStatus,&x.ReputationLabel);err!=nil{return nil,err};if x.PhoneID!=nil{ids,_:=s.Repository.IdentitiesByPhoneID(ctx,*x.PhoneID);for i:=range ids{if ids[i].IsPrimary{x.PublicIdentity=&ids[i];break}}};out=append(out,x)};return out,rows.Err()
}

func(s Service) RescanContacts(ctx context.Context,ownerKey string)(int,error){
 rows,err:=s.Repository.DB.QueryContext(ctx,`SELECT c.id::text,c.local_name,c.normalized_e164 FROM private_contacts c JOIN contact_books b ON b.id=c.contact_book_id WHERE b.owner_key=$1 AND c.normalized_e164 IS NOT NULL`,ownerKey);if err!=nil{return 0,err};defer rows.Close()
 type item struct{id,name,e164 string};items:=[]item{};for rows.Next(){var x item;var name *string;if err:=rows.Scan(&x.id,&name,&x.e164);err!=nil{return 0,err};if name!=nil{x.name=*name};items=append(items,x)}
 changed:=0;for _,x:=range items{n,err:=s.Repository.FindByE164(ctx,x.e164);status:="unknown";var pid any=nil;if err==nil{pid=n.ID;ids,_:=s.Repository.IdentitiesByPhoneID(ctx,n.ID);var primary *Identity;for i:=range ids{if ids[i].IsPrimary{primary=&ids[i];break}};disputed,_:=s.Repository.HasOpenIdentityDispute(ctx,n.ID);status=contactMatch(x.name,primary,disputed)};res,err:=s.Repository.DB.ExecContext(ctx,`UPDATE private_contacts SET phone_number_id=$2,match_status=$3,last_checked_at=NOW(),updated_at=NOW() WHERE id=$1 AND (phone_number_id IS DISTINCT FROM $2 OR match_status::text<>$3)`,x.id,pid,status);if err!=nil{return changed,err};if n,_:=res.RowsAffected();n>0{changed++}}
 return changed,nil
}
