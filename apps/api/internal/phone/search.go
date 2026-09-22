package phone

import ("context";"strings")
type SearchItem struct{E164 string `json:"e164"`;DisplayName *string `json:"display_name,omitempty"`;VerificationStatus string `json:"verification_status"`;SpamScore float64 `json:"spam_score"`;SearchCount int64 `json:"search_count"`}
func (r Repository) Search(ctx context.Context,q,country string,limit int)([]SearchItem,error){
 q=strings.TrimSpace(q);country=strings.ToUpper(strings.TrimSpace(country));if limit<1||limit>50{limit=20}
 rows,err:=r.DB.QueryContext(ctx,`SELECT p.e164,i.display_name,p.verification_status::text,p.spam_score::float8,p.search_count
 FROM phone_numbers p LEFT JOIN LATERAL(SELECT display_name FROM phone_identities WHERE phone_number_id=p.id AND is_public=TRUE ORDER BY is_primary DESC,confidence_score DESC LIMIT 1)i ON TRUE
 WHERE ($1='' OR p.e164 LIKE $1||'%') AND ($2='' OR p.country_code=$2)
 ORDER BY p.search_count DESC,p.last_seen_at DESC LIMIT $3`,q,country,limit);if err!=nil{return nil,err};defer rows.Close()
 out:=make([]SearchItem,0);for rows.Next(){var x SearchItem;if err:=rows.Scan(&x.E164,&x.DisplayName,&x.VerificationStatus,&x.SpamScore,&x.SearchCount);err!=nil{return nil,err};out=append(out,x)};return out,rows.Err()
}
