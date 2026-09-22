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
