package httpserver

import (
 "context"
 "encoding/json"
 "errors"
 "net/http"
 "strings"
 "time"
 "github.com/wuangmmo/sodienthoai-platform/apps/api/internal/phone"
)

type CommunityHandler struct{Service phone.Service}
func subject(r *http.Request)string{return strings.TrimSpace(r.Header.Get("X-User-Subject"))}

func(h CommunityHandler) Follow(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return}
 e164,err:=phone.NormalizeForCountry(r.PathValue("number"),r.URL.Query().Get("country"));if err!=nil{writeJSON(w,400,map[string]string{"error":"invalid_phone_number"});return}
 var in struct{Follow bool `json:"follow"`};if json.NewDecoder(http.MaxBytesReader(w,r.Body,1024)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel();if err:=h.Service.Follow(ctx,sub,e164,in.Follow);err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};writeJSON(w,200,map[string]any{"following":in.Follow})
}
func(h CommunityHandler) Comment(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return}
 e164,err:=phone.NormalizeForCountry(r.PathValue("number"),r.URL.Query().Get("country"));if err!=nil{writeJSON(w,400,map[string]string{"error":"invalid_phone_number"});return}
 var in struct{Body string `json:"body"`};if json.NewDecoder(http.MaxBytesReader(w,r.Body,4096)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel();id,err:=h.Service.Comment(ctx,sub,e164,in.Body);if errors.Is(err,phone.ErrInvalidComment){writeJSON(w,400,map[string]string{"error":"invalid_comment"});return};if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};writeJSON(w,202,map[string]string{"status":"pending","id":id})
}


func(h CommunityHandler) ListComments(w http.ResponseWriter,r *http.Request){
 e164,err:=phone.NormalizeForCountry(r.PathValue("number"),r.URL.Query().Get("country"));if err!=nil{writeJSON(w,400,map[string]string{"error":"invalid_phone_number"});return};ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel();items,err:=h.Service.PublicComments(ctx,e164,30);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};writeJSON(w,200,map[string]any{"data":items})
}
func(h CommunityHandler) Helpful(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return};ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel();if err:=h.Service.HelpfulComment(ctx,sub,r.PathValue("id"));err!=nil{writeJSON(w,400,map[string]string{"error":"vote_failed"});return};writeJSON(w,200,map[string]string{"status":"ok"})
}


func(h CommunityHandler) ReportComment(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return};var in struct{Reason string `json:"reason"`};if json.NewDecoder(http.MaxBytesReader(w,r.Body,2048)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return};ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel();if err:=h.Service.ReportComment(ctx,sub,r.PathValue("id"),in.Reason);err!=nil{writeJSON(w,400,map[string]string{"error":"report_failed"});return};writeJSON(w,202,map[string]string{"status":"accepted"})
}
func(h CommunityHandler) Follows(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return};ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel();items,err:=h.Service.FollowedPhones(ctx,sub);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};writeJSON(w,200,map[string]any{"data":items})
}
func(h CommunityHandler) Notifications(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return};ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel();items,err:=h.Service.Notifications(ctx,sub);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};writeJSON(w,200,map[string]any{"data":items})
}


func(h CommunityHandler) MarkNotificationRead(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return};ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel();if err:=h.Service.MarkNotificationRead(ctx,sub,r.PathValue("id"));err!=nil{writeJSON(w,404,map[string]string{"error":"notification_not_found"});return};writeJSON(w,200,map[string]string{"status":"read"})
}
