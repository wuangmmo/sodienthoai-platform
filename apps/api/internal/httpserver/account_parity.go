package httpserver

import (
 "context"
 "encoding/json"
 "errors"
 "net/http"
 "time"

 "github.com/wuangmmo/sodienthoai-platform/apps/api/internal/phone"
)

type AccountParityHandler struct{ Service phone.Service }

func (h AccountParityHandler) Preferences(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel()
 items,err:=h.Service.PhonePreferences(ctx,sub,r.URL.Query().Get("disposition"));if errors.Is(err,phone.ErrInvalidPreference){writeJSON(w,400,map[string]string{"error":"invalid_disposition"});return};if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return}
 writeJSON(w,200,map[string]any{"data":items})
}

func (h AccountParityHandler) SetPreference(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return}
 e164,err:=phone.NormalizeForCountry(r.PathValue("number"),r.URL.Query().Get("country"));if err!=nil{writeJSON(w,400,map[string]string{"error":"invalid_phone_number"});return}
 var in struct{Disposition string `json:"disposition"`};if json.NewDecoder(http.MaxBytesReader(w,r.Body,1024)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel();err=h.Service.SetPhonePreference(ctx,sub,e164,in.Disposition)
 if errors.Is(err,phone.ErrInvalidPreference){writeJSON(w,400,map[string]string{"error":"invalid_disposition"});return};if phone.IsNotFound(err){writeJSON(w,404,map[string]string{"error":"phone_number_not_found"});return};if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return}
 writeJSON(w,200,map[string]string{"status":"saved"})
}

func (h AccountParityHandler) DeletePreference(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return}
 e164,err:=phone.NormalizeForCountry(r.PathValue("number"),r.URL.Query().Get("country"));if err!=nil{writeJSON(w,400,map[string]string{"error":"invalid_phone_number"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel();err=h.Service.DeletePhonePreference(ctx,sub,e164);if phone.IsNotFound(err){writeJSON(w,404,map[string]string{"error":"phone_number_not_found"});return};if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};writeJSON(w,200,map[string]string{"status":"deleted"})
}

func (h AccountParityHandler) History(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel();items,err:=h.Service.UserLookupHistory(ctx,sub);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};writeJSON(w,200,map[string]any{"data":items})
}

func (h AccountParityHandler) DeleteHistory(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel();if err:=h.Service.DeleteUserLookupHistory(ctx,sub);err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};writeJSON(w,200,map[string]string{"status":"deleted"})
}

func (h AccountParityHandler) Appeal(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return}
 e164,err:=phone.NormalizeForCountry(r.PathValue("number"),r.URL.Query().Get("country"));if err!=nil{writeJSON(w,400,map[string]string{"error":"invalid_phone_number"});return}
 var in phone.AppealInput;if json.NewDecoder(http.MaxBytesReader(w,r.Body,8192)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel();id,err:=h.Service.CreateAppeal(ctx,sub,e164,in)
 if errors.Is(err,phone.ErrInvalidAppeal){writeJSON(w,400,map[string]string{"error":"invalid_appeal"});return};if errors.Is(err,phone.ErrDuplicateAppeal){writeJSON(w,409,map[string]string{"error":"pending_appeal_exists"});return};if phone.IsNotFound(err){writeJSON(w,404,map[string]string{"error":"phone_number_not_found"});return};if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return}
 writeJSON(w,202,map[string]string{"status":"pending","id":id})
}
