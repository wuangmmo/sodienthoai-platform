package httpserver

import (
 "context"
 "encoding/json"
 "errors"
 "net/http"
 "io"
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


func (h ContactHandler) ImportFile(w http.ResponseWriter,r *http.Request){
 owner:=strings.TrimSpace(r.Header.Get("X-Contact-Owner"));if owner==""{writeJSON(w,401,map[string]string{"error":"contact_owner_required"});return}
 if err:=r.ParseMultipartForm(5<<20);err!=nil{writeJSON(w,400,map[string]string{"error":"invalid_upload"});return};f,hdr,err:=r.FormFile("file");if err!=nil{writeJSON(w,400,map[string]string{"error":"file_required"});return};defer f.Close();data,err:=io.ReadAll(io.LimitReader(f,5<<20));if err!=nil{writeJSON(w,400,map[string]string{"error":"invalid_upload"});return}
 var items []phone.ContactImportItem;name:=strings.ToLower(hdr.Filename);if strings.HasSuffix(name,".csv"){items,err=phone.ParseContactsCSV(data)}else if strings.HasSuffix(name,".vcf")||strings.HasSuffix(name,".vcard"){items,err=phone.ParseContactsVCard(data)}else{writeJSON(w,400,map[string]string{"error":"unsupported_file"});return};if err!=nil{writeJSON(w,400,map[string]string{"error":"invalid_contact_file"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),15*time.Second);defer cancel();result,err:=h.Service.ImportContacts(ctx,owner,items);if err!=nil{writeJSON(w,400,map[string]string{"error":"contact_import_failed"});return};writeJSON(w,201,map[string]any{"data":result})
}

func (h ContactHandler) Alerts(w http.ResponseWriter,r *http.Request){
 owner:=strings.TrimSpace(r.Header.Get("X-Contact-Owner"));if owner==""{writeJSON(w,401,map[string]string{"error":"contact_owner_required"});return}
 ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel();rows,err:=h.Service.Repository.DB.QueryContext(ctx,`SELECT e.id::text,e.contact_id::text,e.old_match_status::text,e.new_match_status::text,e.created_at FROM contact_change_events e JOIN private_contacts c ON c.id=e.contact_id JOIN contact_books b ON b.id=c.contact_book_id WHERE b.owner_key=$1 AND e.seen_at IS NULL ORDER BY e.created_at DESC LIMIT 100`,owner);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};defer rows.Close();items:=[]map[string]any{};for rows.Next(){var id,cid,oldStatus,newStatus string;var created time.Time;if rows.Scan(&id,&cid,&oldStatus,&newStatus,&created)==nil{items=append(items,map[string]any{"id":id,"contact_id":cid,"old_status":oldStatus,"new_status":newStatus,"created_at":created})}};writeJSON(w,200,map[string]any{"data":items})
}
