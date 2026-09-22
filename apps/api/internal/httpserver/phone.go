package httpserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"net/http"
	"time"

	"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/phone"
)

type PhoneHandler struct{ Service phone.Service }

func (h PhoneHandler) Get(w http.ResponseWriter,r *http.Request){
	e164,err:=phone.NormalizeForCountry(r.PathValue("number"),r.URL.Query().Get("country"))
	if err!=nil { writeJSON(w,http.StatusBadRequest,map[string]string{"error":"invalid_phone_number"});return }
	ctx,cancel:=context.WithTimeout(r.Context(),2*time.Second);defer cancel()
	result,err:=h.Service.Lookup(ctx,e164)
	if phone.IsNotFound(err){writeJSON(w,http.StatusNotFound,map[string]any{"error":"phone_number_not_found","number":e164,"identified":false});return}
	if err!=nil{writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"});return}
	writeJSON(w,http.StatusOK,map[string]any{"data":result})
}

func (h PhoneHandler) Report(w http.ResponseWriter,r *http.Request){
 e164,err:=phone.NormalizeForCountry(r.PathValue("number"),r.URL.Query().Get("country"));if err!=nil{writeJSON(w,400,map[string]string{"error":"invalid_phone_number"});return}
 var in phone.ReportInput;if json.NewDecoder(http.MaxBytesReader(w,r.Body,4096)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return}
 sum:=sha256.Sum256([]byte(strings.TrimSpace(r.RemoteAddr)));ctx,cancel:=context.WithTimeout(r.Context(),2*time.Second);defer cancel()
 err=h.Service.Report(ctx,e164,in,hex.EncodeToString(sum[:]));if errors.Is(err,phone.ErrInvalidReport){writeJSON(w,400,map[string]string{"error":"invalid_report"});return};if phone.IsNotFound(err){writeJSON(w,404,map[string]string{"error":"phone_number_not_found"});return};if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return}
 writeJSON(w,202,map[string]string{"status":"pending_review"})
}

func (h PhoneHandler) Claim(w http.ResponseWriter,r *http.Request){
 e164,err:=phone.NormalizeForCountry(r.PathValue("number"),r.URL.Query().Get("country"));if err!=nil{writeJSON(w,400,map[string]string{"error":"invalid_phone_number"});return}
 var in phone.ClaimInput;if json.NewDecoder(http.MaxBytesReader(w,r.Body,8192)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),2*time.Second);defer cancel();err=h.Service.Claim(ctx,e164,in)
 if errors.Is(err,phone.ErrInvalidClaim){writeJSON(w,400,map[string]string{"error":"invalid_claim"});return};if phone.IsNotFound(err){writeJSON(w,404,map[string]string{"error":"phone_number_not_found"});return};if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return}
 writeJSON(w,202,map[string]string{"status":"pending_verification"})
}
