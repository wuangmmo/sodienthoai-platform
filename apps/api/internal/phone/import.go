package phone
import("context";"database/sql")
type ImportRow struct{CountryCode string;CallingCode string;NationalNumber string;E164 string;NumberType string}
func(r Repository)ImportRows(ctx context.Context,source string,rows []ImportRow)(accepted,rejected int64,err error){
 tx,err:=r.DB.BeginTx(ctx,nil);if err!=nil{return 0,0,err};defer tx.Rollback()
 var batch string;if err=tx.QueryRowContext(ctx,"INSERT INTO phone_import_batches(source_name,total_rows,status) VALUES($1,$2,'running') RETURNING id::text",source,len(rows)).Scan(&batch);err!=nil{return}
 stmt,err:=tx.PrepareContext(ctx,`INSERT INTO phone_numbers(country_code,calling_code,national_number,e164,number_type,seo_status) VALUES($1,$2,$3,$4,NULLIF($5,''),'noindex') ON CONFLICT(e164) DO UPDATE SET last_seen_at=NOW()`);if err!=nil{return};defer stmt.Close()
 for _,x:=range rows{if _,e:=stmt.ExecContext(ctx,x.CountryCode,x.CallingCode,x.NationalNumber,x.E164,x.NumberType);e!=nil{rejected++}else{accepted++}}
 _,err=tx.ExecContext(ctx,"UPDATE phone_import_batches SET accepted_rows=$2,rejected_rows=$3,status='completed',completed_at=NOW() WHERE id=$1",batch,accepted,rejected);if err!=nil{return};err=tx.Commit();return
}
var _ *sql.Tx
