package httpserver

import (
 "context"
 "net/http"
 "strconv"
 "time"
)

func adminLimit(r *http.Request) int { n,_:=strconv.Atoi(r.URL.Query().Get("limit"));if n<1||n>100{return 50};return n }

func(a AdminHandler) Appeals(w http.ResponseWriter,r *http.Request){
 if !a.authorized(r){writeJSON(w,401,map[string]string{"error":"unauthorized"});return};ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel()
 rows,err:=a.DB.QueryContext(ctx,`SELECT a.id::text,p.e164,a.reason,a.statement,a.status,a.created_at FROM phone_appeals a JOIN phone_numbers p ON p.id=a.phone_number_id ORDER BY CASE WHEN a.status='pending' THEN 0 ELSE 1 END,a.created_at DESC LIMIT $1`,adminLimit(r));if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};defer rows.Close()
 items:=[]map[string]any{};for rows.Next(){var id,e164,reason,statement,status string;var created time.Time;if rows.Scan(&id,&e164,&reason,&statement,&status,&created)==nil{items=append(items,map[string]any{"id":id,"e164":e164,"reason":reason,"statement":statement,"status":status,"created_at":created})}};writeJSON(w,200,map[string]any{"data":items})
}
func(a AdminHandler) BusinessVerifications(w http.ResponseWriter,r *http.Request){
 if !a.authorized(r){writeJSON(w,401,map[string]string{"error":"unauthorized"});return};ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel()
 rows,err:=a.DB.QueryContext(ctx,`SELECT v.id::text,b.id::text,b.display_name,v.method,v.status,v.statement,v.created_at FROM business_verification_requests v JOIN businesses b ON b.id=v.business_id ORDER BY CASE WHEN v.status='pending' THEN 0 ELSE 1 END,v.created_at DESC LIMIT $1`,adminLimit(r));if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};defer rows.Close()
 items:=[]map[string]any{};for rows.Next(){var id,bid,name,method,status string;var statement *string;var created time.Time;if rows.Scan(&id,&bid,&name,&method,&status,&statement,&created)==nil{items=append(items,map[string]any{"id":id,"business_id":bid,"business_name":name,"method":method,"status":status,"statement":statement,"created_at":created})}};writeJSON(w,200,map[string]any{"data":items})
}
func(a AdminHandler) BusinessReviews(w http.ResponseWriter,r *http.Request){
 if !a.authorized(r){writeJSON(w,401,map[string]string{"error":"unauthorized"});return};ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel()
 rows,err:=a.DB.QueryContext(ctx,`SELECT rv.id::text,b.id::text,b.display_name,rv.rating,rv.body,rv.status,rv.created_at FROM business_reviews rv JOIN businesses b ON b.id=rv.business_id ORDER BY CASE WHEN rv.status='pending' THEN 0 ELSE 1 END,rv.created_at DESC LIMIT $1`,adminLimit(r));if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};defer rows.Close()
 items:=[]map[string]any{};for rows.Next(){var id,bid,name,status string;var rating int;var body *string;var created time.Time;if rows.Scan(&id,&bid,&name,&rating,&body,&status,&created)==nil{items=append(items,map[string]any{"id":id,"business_id":bid,"business_name":name,"rating":rating,"body":body,"status":status,"created_at":created})}};writeJSON(w,200,map[string]any{"data":items})
}
