package httpserver
import("context";"net/http";"strconv";"strings";"time";"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/phone";searchpkg "github.com/wuangmmo/sodienthoai-platform/apps/api/internal/search")
type SearchHandler struct{Repository phone.Repository;Search *searchpkg.Client}
func(h SearchHandler)Get(w http.ResponseWriter,r *http.Request){limit,_:=strconv.Atoi(r.URL.Query().Get("limit"));q:=strings.TrimSpace(r.URL.Query().Get("q"));country:=strings.TrimSpace(r.URL.Query().Get("country"));ctx,cancel:=context.WithTimeout(r.Context(),2*time.Second);defer cancel()
 if h.Search!=nil&&q!=""{if items,err:=h.Search.SearchPhones(ctx,q,country,limit);err==nil{writeJSON(w,200,map[string]any{"data":items,"engine":"opensearch"});return}}
 items,err:=h.Repository.Search(ctx,q,country,limit);if err!=nil{writeJSON(w,500,map[string]string{"error":"internal_error"});return};writeJSON(w,200,map[string]any{"data":items,"engine":"postgres"})}
