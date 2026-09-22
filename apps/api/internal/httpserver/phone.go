package httpserver

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"net/http"
	"net"
	"time"

	"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/phone"
)

type PhoneHandler struct{ Service phone.Service; ReporterHashSecret string }

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
 host:=strings.TrimSpace(r.RemoteAddr);if h,_,splitErr:=net.SplitHostPort(host);splitErr==nil{host=h};mac:=hmac.New(sha256.New,[]byte(h.ReporterHashSecret));_,_=mac.Write([]byte(host));reporterHash:=hex.EncodeToString(mac.Sum(nil));ctx,cancel:=context.WithTimeout(r.Context(),2*time.Second);defer cancel()
 err=h.Service.Report(ctx,e164,in,reporterHash);if errors.Is(err,phone.ErrInvalidReport){writeJSON(w,400,map[string]string{"error":"invalid_report"});return};if errors.Is(err,phone.ErrDuplicateReport){writeJSON(w,409,map[string]string{"error":"duplicate_report","message":"report_already_submitted_recently"});return};if phone.IsNotFound(err){writeJSON(w,404,map[string]string{"error":"phone_number_not_found"});return};if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return}
 writeJSON(w,202,map[string]string{"status":"pending_review"})
}

func (h PhoneHandler) Claim(w http.ResponseWriter,r *http.Request){
 e164,err:=phone.NormalizeForCountry(r.PathValue("number"),r.URL.Query().Get("country"));if err!=nil{writeJSON(w,400,map[string]string{"error":"invalid_phone_number"});return}
 var in phone.ClaimInput;if json.NewDecoder(http.MaxBytesReader(w,r.Body,8192)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),2*time.Second);defer cancel();err=h.Service.Claim(ctx,e164,in)
 if errors.Is(err,phone.ErrInvalidClaim){writeJSON(w,400,map[string]string{"error":"invalid_claim"});return};if phone.IsNotFound(err){writeJSON(w,404,map[string]string{"error":"phone_number_not_found"});return};if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return}
 writeJSON(w,202,map[string]string{"status":"pending_verification"})
}


func (h PhoneHandler) Profile(w http.ResponseWriter,r *http.Request){
	e164,err:=phone.NormalizeForCountry(r.PathValue("number"),r.URL.Query().Get("country"))
	if err!=nil{writeJSON(w,http.StatusBadRequest,map[string]string{"error":"invalid_phone_number"});return}
	ctx,cancel:=context.WithTimeout(r.Context(),2*time.Second);defer cancel()
	profile,err:=h.Service.Profile(ctx,e164)
	if err!=nil{writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"});return}
	writeJSON(w,http.StatusOK,map[string]any{"data":profile})
}
