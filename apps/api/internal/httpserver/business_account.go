package httpserver

import (
 "context"
 "net/http"
 "time"
)

func(h BusinessHandler) Mine(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel();uid,err:=h.Service.EnsureUser(ctx,sub);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return}
 rows,err:=h.Service.Repository.DB.QueryContext(ctx,`SELECT b.id::text,b.slug,b.display_name,b.verification_status,o.role,o.status,b.updated_at FROM businesses b JOIN business_ownerships o ON o.business_id=b.id WHERE o.user_id=$1 ORDER BY b.updated_at DESC`,uid);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};defer rows.Close()
 items:=[]map[string]any{};for rows.Next(){var id,slug,name,verification,role,ownership string;var updated time.Time;if rows.Scan(&id,&slug,&name,&verification,&role,&ownership,&updated)==nil{items=append(items,map[string]any{"id":id,"slug":slug,"display_name":name,"verification_status":verification,"role":role,"ownership_status":ownership,"updated_at":updated})}}
 writeJSON(w,200,map[string]any{"data":items})
}

func(h BusinessHandler) MyVerifications(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel();uid,err:=h.Service.EnsureUser(ctx,sub);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return}
 rows,err:=h.Service.Repository.DB.QueryContext(ctx,`SELECT v.id::text,v.business_id::text,b.display_name,v.method,v.status,v.created_at,v.reviewed_at FROM business_verification_requests v JOIN businesses b ON b.id=v.business_id WHERE v.user_id=$1 ORDER BY v.created_at DESC LIMIT 100`,uid);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};defer rows.Close()
 items:=[]map[string]any{};for rows.Next(){var id,bid,name,method,status string;var created time.Time;var reviewed *time.Time;if rows.Scan(&id,&bid,&name,&method,&status,&created,&reviewed)==nil{items=append(items,map[string]any{"id":id,"business_id":bid,"display_name":name,"method":method,"status":status,"created_at":created,"reviewed_at":reviewed})}}
 writeJSON(w,200,map[string]any{"data":items})
}
