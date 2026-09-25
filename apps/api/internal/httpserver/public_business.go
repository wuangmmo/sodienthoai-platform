package httpserver

import (
 "context"
 "net/http"
 "time"

 "github.com/wuangmmo/sodienthoai-platform/apps/api/internal/phone"
)

type PublicBusinessHandler struct{Service phone.Service}

func(h PublicBusinessHandler) Get(w http.ResponseWriter,r *http.Request){
 slug:=r.PathValue("slug");ctx,cancel:=context.WithTimeout(r.Context(),4*time.Second);defer cancel()
 row:=h.Service.Repository.DB.QueryRowContext(ctx,`SELECT id::text,slug,display_name,description,website_url,updated_at FROM businesses WHERE slug=$1 AND verification_status='verified'`,slug)
 var id,s,name string;var description,website *string;var updated time.Time
 if err:=row.Scan(&id,&s,&name,&description,&website,&updated);err!=nil{writeJSON(w,404,map[string]string{"error":"business_not_found"});return}
 branches:=[]map[string]any{};rows,err:=h.Service.Repository.DB.QueryContext(ctx,`SELECT id::text,name,address_text,region_code,latitude::text,longitude::text,is_primary FROM business_branches WHERE business_id=$1 ORDER BY is_primary DESC,created_at`,id);if err==nil{defer rows.Close();for rows.Next(){var bid,bname string;var address,region,lat,lng *string;var primary bool;if rows.Scan(&bid,&bname,&address,&region,&lat,&lng,&primary)==nil{branches=append(branches,map[string]any{"id":bid,"name":bname,"address":address,"region_code":region,"latitude":lat,"longitude":lng,"is_primary":primary})}}}
 phones:=[]string{};prows,err:=h.Service.Repository.DB.QueryContext(ctx,`SELECT p.e164 FROM business_phone_links l JOIN phone_numbers p ON p.id=l.phone_number_id WHERE l.business_id=$1 AND l.is_public=TRUE ORDER BY l.created_at`,id);if err==nil{defer prows.Close();for prows.Next(){var e164 string;if prows.Scan(&e164)==nil{phones=append(phones,e164)}}}
 var reviewCount int64;var rating *float64;_ = h.Service.Repository.DB.QueryRowContext(ctx,`SELECT COUNT(*),AVG(rating)::float8 FROM business_reviews WHERE business_id=$1 AND status='approved'`,id).Scan(&reviewCount,&rating)
 writeJSON(w,200,map[string]any{"data":map[string]any{"id":id,"slug":s,"display_name":name,"description":description,"website_url":website,"verified":true,"phones":phones,"branches":branches,"review_count":reviewCount,"rating":rating,"updated_at":updated}})
}
