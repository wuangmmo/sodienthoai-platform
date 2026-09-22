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

type ContactHandler struct{Service phone.Service}
type contactImportRequest struct{Contacts []phone.ContactImportItem `json:"contacts"`}

func (h ContactHandler) Import(w http.ResponseWriter,r *http.Request){
 owner:=strings.TrimSpace(r.Header.Get("X-Contact-Owner"));if owner==""{writeJSON(w,401,map[string]string{"error":"contact_owner_required"});return}
 var in contactImportRequest;if json.NewDecoder(http.MaxBytesReader(w,r.Body,2<<20)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),10*time.Second);defer cancel();result,err:=h.Service.ImportContacts(ctx,owner,in.Contacts)
 if err!=nil{writeJSON(w,400,map[string]string{"error":"contact_import_failed"});return};writeJSON(w,201,map[string]any{"data":result})
}


func (h ContactHandler) Action(w http.ResponseWriter,r *http.Request){
 owner:=strings.TrimSpace(r.Header.Get("X-Contact-Owner"));if owner==""{writeJSON(w,401,map[string]string{"error":"contact_owner_required"});return}
 var in struct{Action string `json:"action"`};if json.NewDecoder(http.MaxBytesReader(w,r.Body,4096)).Decode(&in)!=nil{writeJSON(w,400,map[string]string{"error":"invalid_request"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel();id,err:=h.Service.ContactAction(ctx,owner,r.PathValue("id"),in.Action)
 if errors.Is(err,phone.ErrInvalidContactAction){writeJSON(w,400,map[string]string{"error":"invalid_action"});return};if err!=nil{writeJSON(w,404,map[string]string{"error":"contact_not_found"});return};writeJSON(w,202,map[string]string{"status":"accepted","action_id":id})
}

func (h ContactHandler) DeleteAll(w http.ResponseWriter,r *http.Request){
 owner:=strings.TrimSpace(r.Header.Get("X-Contact-Owner"));if owner==""{writeJSON(w,401,map[string]string{"error":"contact_owner_required"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel();if err:=h.Service.DeleteContactBook(ctx,owner);err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};w.WriteHeader(http.StatusNoContent)
}


func (h ContactHandler) List(w http.ResponseWriter,r *http.Request){
 owner:=strings.TrimSpace(r.Header.Get("X-Contact-Owner"));if owner==""{writeJSON(w,401,map[string]string{"error":"contact_owner_required"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel();items,err:=h.Service.ListContacts(ctx,owner,r.URL.Query().Get("status"),100);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};writeJSON(w,200,map[string]any{"data":items})
}
func (h ContactHandler) Rescan(w http.ResponseWriter,r *http.Request){
 owner:=strings.TrimSpace(r.Header.Get("X-Contact-Owner"));if owner==""{writeJSON(w,401,map[string]string{"error":"contact_owner_required"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),15*time.Second);defer cancel();changed,err:=h.Service.RescanContacts(ctx,owner);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};writeJSON(w,200,map[string]any{"changed":changed})
}
