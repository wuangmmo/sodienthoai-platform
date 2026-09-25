package httpserver

import (
 "context"
 "encoding/json"
 "errors"
 "net/http"
 "time"

 "github.com/wuangmmo/sodienthoai-platform/apps/api/internal/phone"
)

type BusinessHandler struct{Service phone.Service}

func(h BusinessHandler) Create(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return}
 var in phone.BusinessCreateInput;if json.NewDecoder(http.MaxBytesReader(w,r.Body,16384)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel();id,err:=h.Service.CreateBusiness(ctx,sub,in)
 if errors.Is(err,phone.ErrInvalidBusiness){writeJSON(w,400,map[string]string{"error":"invalid_business"});return};if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};writeJSON(w,201,map[string]string{"id":id,"status":"unverified"})
}

func(h BusinessHandler) Verification(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return}
 var in phone.BusinessVerificationInput;if json.NewDecoder(http.MaxBytesReader(w,r.Body,16384)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel();id,err:=h.Service.RequestBusinessVerification(ctx,sub,r.PathValue("id"),in)
 if errors.Is(err,phone.ErrInvalidBusiness){writeJSON(w,403,map[string]string{"error":"business_access_denied"});return};if errors.Is(err,phone.ErrDuplicateBusinessVerification){writeJSON(w,409,map[string]string{"error":"pending_verification_exists"});return};if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};writeJSON(w,202,map[string]string{"id":id,"status":"pending"})
}

func(h BusinessHandler) Review(w http.ResponseWriter,r *http.Request){
 sub:=subject(r);if sub==""{writeJSON(w,401,map[string]string{"error":"user_required"});return}
 var in phone.BusinessReviewInput;if json.NewDecoder(http.MaxBytesReader(w,r.Body,8192)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel();id,err:=h.Service.CreateBusinessReview(ctx,sub,r.PathValue("id"),in)
 if errors.Is(err,phone.ErrInvalidBusinessReview){writeJSON(w,400,map[string]string{"error":"invalid_review"});return};if errors.Is(err,phone.ErrDuplicateBusinessReview){writeJSON(w,409,map[string]string{"error":"active_review_exists"});return};if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};writeJSON(w,202,map[string]string{"id":id,"status":"pending"})
}
