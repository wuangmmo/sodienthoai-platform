package httpserver

import (
 "context"
 "encoding/json"
 "net/http"
 "strings"
 "time"
 "github.com/wuangmmo/sodienthoai-platform/apps/api/internal/phone"
)

type ContactHandler struct{Service phone.Service}
type contactImportRequest struct{Contacts []phone.ContactImportItem `json:"contacts"`}

func (h ContactHandler) Import(w http.ResponseWriter,r *http.Request){
 owner:=strings.TrimSpace(r.Header.Get("X-Contact-Owner"));if owner==""{writeJSON(w,401,map[string]string{"error":"contact_owner_required"});return}
 var in contactImportRequest;if json.NewDecoder(http.MaxBytesReader(w,r.Body,2<<20)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),10*time.Second);defer cancel();result,err:=h.Service.ImportContacts(ctx,owner,in.Contacts)
 if err!=nil{writeJSON(w,400,map[string]string{"error":"contact_import_failed"});return};writeJSON(w,201,map[string]any{"data":result})
}
