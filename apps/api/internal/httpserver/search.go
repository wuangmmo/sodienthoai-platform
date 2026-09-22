package httpserver
import("context";"net/http";"strconv";"time";"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/phone")
type SearchHandler struct{Repository phone.Repository}
func(h SearchHandler)Get(w http.ResponseWriter,r *http.Request){limit,_:=strconv.Atoi(r.URL.Query().Get("limit"));ctx,cancel:=context.WithTimeout(r.Context(),2*time.Second);defer cancel();items,err:=h.Repository.Search(ctx,r.URL.Query().Get("q"),r.URL.Query().Get("country"),limit);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};writeJSON(w,200,map[string]any{"data":items})}
