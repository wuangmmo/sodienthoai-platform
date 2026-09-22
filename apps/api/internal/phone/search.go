package phone

import ("context";"encoding/base64";"fmt";"strconv";"strings")
type SearchItem struct{E164 string `json:"e164"`;DisplayName *string `json:"display_name,omitempty"`;VerificationStatus string `json:"verification_status"`;SpamScore float64 `json:"spam_score"`;SearchCount int64 `json:"search_count"`}
type SearchPage struct{Items []SearchItem `json:"items"`;NextCursor string `json:"next_cursor,omitempty"`}
func encodeCursor(count int64,e164 string)string{return base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%d|%s",count,e164)))}
func decodeCursor(raw string)(int64,string){if raw==""{return -1,""};b,err:=base64.RawURLEncoding.DecodeString(raw);if err!=nil{return -1,""};parts:=strings.SplitN(string(b),"|",2);if len(parts)!=2{return -1,""};n,err:=strconv.ParseInt(parts[0],10,64);if err!=nil{return -1,""};return n,parts[1]}
func (r Repository) SearchPage(ctx context.Context,q,country,cursor string,limit int)(SearchPage,error){
 q=strings.TrimSpace(q);country=strings.ToUpper(strings.TrimSpace(country));if limit<1||limit>50{limit=20};cc,ce:=decodeCursor(cursor)
 rows,err:=r.DB.QueryContext(ctx,`SELECT p.e164,i.display_name,p.verification_status::text,p.spam_score::float8,p.search_count
 FROM phone_numbers p LEFT JOIN LATERAL(SELECT display_name FROM phone_identities WHERE phone_number_id=p.id AND is_public=TRUE ORDER BY is_primary DESC,confidence_score DESC LIMIT 1)i ON TRUE
 WHERE ($1='' OR p.e164 LIKE $1||'%') AND ($2='' OR p.country_code=$2) AND ($3::bigint<0 OR (p.search_count,p.e164)<($3,$4))
 ORDER BY p.search_count DESC,p.e164 DESC LIMIT $5`,q,country,cc,ce,limit+1);if err!=nil{return SearchPage{},err};defer rows.Close()
 out:=make([]SearchItem,0,limit+1);for rows.Next(){var x SearchItem;if err:=rows.Scan(&x.E164,&x.DisplayName,&x.VerificationStatus,&x.SpamScore,&x.SearchCount);err!=nil{return SearchPage{},err};out=append(out,x)};if err:=rows.Err();err!=nil{return SearchPage{},err}
 page:=SearchPage{Items:out};if len(out)>limit{last:=out[limit-1];page.Items=out[:limit];page.NextCursor=encodeCursor(last.SearchCount,last.E164)};return page,nil
}
func (r Repository) Search(ctx context.Context,q,country string,limit int)([]SearchItem,error){p,err:=r.SearchPage(ctx,q,country,"",limit);return p.Items,err}
