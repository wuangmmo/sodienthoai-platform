package phone

import (
 "context"
 "errors"
 "strings"
)

var ErrInvalidBusiness = errors.New("invalid business")
var ErrInvalidBusinessReview = errors.New("invalid business review")
var ErrDuplicateBusinessReview = errors.New("active business review already exists")
var ErrDuplicateBusinessVerification = errors.New("pending business verification already exists")

type BusinessCreateInput struct {
 Slug string `json:"slug"`
 LegalName string `json:"legal_name"`
 DisplayName string `json:"display_name"`
 Description string `json:"description"`
 WebsiteURL string `json:"website_url"`
}
type BusinessVerificationInput struct {
 Method string `json:"method"`
 Statement string `json:"statement"`
 EvidenceRef string `json:"evidence_ref"`
}
type BusinessReviewInput struct {
 Rating int `json:"rating"`
 Body string `json:"body"`
 BranchID string `json:"branch_id"`
}

func validSlug(v string) bool {
 if len(v)<2||len(v)>180{return false}
 for _,c:=range v { if !(c>='a'&&c<='z'||c>='0'&&c<='9'||c=='-'){return false} }
 return !strings.HasPrefix(v,"-")&&!strings.HasSuffix(v,"-")&&!strings.Contains(v,"--")
}

func (s Service) CreateBusiness(ctx context.Context,subject string,in BusinessCreateInput)(string,error){
 in.Slug=strings.ToLower(strings.TrimSpace(in.Slug));in.LegalName=strings.TrimSpace(in.LegalName);in.DisplayName=strings.TrimSpace(in.DisplayName);in.Description=strings.TrimSpace(in.Description);in.WebsiteURL=strings.TrimSpace(in.WebsiteURL)
 if !validSlug(in.Slug)||len(in.LegalName)<2||len(in.LegalName)>200||len(in.DisplayName)<2||len(in.DisplayName)>200||len(in.Description)>4000||len(in.WebsiteURL)>1000{return "",ErrInvalidBusiness}
 uid,err:=s.EnsureUser(ctx,subject);if err!=nil{return "",err}
 tx,err:=s.Repository.DB.BeginTx(ctx,nil);if err!=nil{return "",err};defer tx.Rollback()
 var id string
 err=tx.QueryRowContext(ctx,`INSERT INTO businesses(slug,legal_name,display_name,description,website_url,created_by) VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,''),$6) RETURNING id::text`,in.Slug,in.LegalName,in.DisplayName,in.Description,in.WebsiteURL,uid).Scan(&id);if err!=nil{return "",err}
 if _,err=tx.ExecContext(ctx,`INSERT INTO business_ownerships(business_id,user_id,role,status) VALUES($1,$2,'owner','pending')`,id,uid);err!=nil{return "",err}
 return id,tx.Commit()
}

func (s Service) RequestBusinessVerification(ctx context.Context,subject,businessID string,in BusinessVerificationInput)(string,error){
 in.Method=strings.ToLower(strings.TrimSpace(in.Method));in.Statement=strings.TrimSpace(in.Statement);in.EvidenceRef=strings.TrimSpace(in.EvidenceRef)
 allowed:=map[string]bool{"phone":true,"email":true,"document":true,"manual":true}
 if !allowed[in.Method]||len(in.Statement)>4000||len(in.EvidenceRef)>1000{return "",ErrInvalidBusiness}
 uid,err:=s.EnsureUser(ctx,subject);if err!=nil{return "",err}
 var owns bool;if err=s.Repository.DB.QueryRowContext(ctx,`SELECT EXISTS(SELECT 1 FROM business_ownerships WHERE business_id=$1 AND user_id=$2 AND status IN ('pending','verified'))`,businessID,uid).Scan(&owns);err!=nil{return "",err};if !owns{return "",ErrInvalidBusiness}
 var id string;err=s.Repository.DB.QueryRowContext(ctx,`INSERT INTO business_verification_requests(business_id,user_id,method,statement,evidence_ref) VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,'')) RETURNING id::text`,businessID,uid,in.Method,in.Statement,in.EvidenceRef).Scan(&id)
 if err!=nil&&strings.Contains(strings.ToLower(err.Error()),"idx_business_verification_one_pending"){return "",ErrDuplicateBusinessVerification};return id,err
}

func (s Service) CreateBusinessReview(ctx context.Context,subject,businessID string,in BusinessReviewInput)(string,error){
 in.Body=strings.TrimSpace(in.Body);in.BranchID=strings.TrimSpace(in.BranchID)
 if in.Rating<1||in.Rating>5||len(in.Body)>2000{return "",ErrInvalidBusinessReview}
 uid,err:=s.EnsureUser(ctx,subject);if err!=nil{return "",err}
 var eligible bool;err=s.Repository.DB.QueryRowContext(ctx,`SELECT EXISTS(SELECT 1 FROM businesses WHERE id=$1 AND verification_status=\'verified\')`,businessID).Scan(&eligible);if err!=nil{return "",err};if !eligible{return "",ErrInvalidBusinessReview}
 if in.BranchID!=""{var branchOK bool;err=s.Repository.DB.QueryRowContext(ctx,`SELECT EXISTS(SELECT 1 FROM business_branches WHERE id=$1 AND business_id=$2)`,in.BranchID,businessID).Scan(&branchOK);if err!=nil{return "",err};if !branchOK{return "",ErrInvalidBusinessReview}}
 var id string
 err=s.Repository.DB.QueryRowContext(ctx,`INSERT INTO business_reviews(business_id,branch_id,user_id,rating,body) VALUES($1,NULLIF($2,\'\')::uuid,$3,$4,NULLIF($5,\'\')) RETURNING id::text`,businessID,in.BranchID,uid,in.Rating,in.Body).Scan(&id)
 if err!=nil&&strings.Contains(strings.ToLower(err.Error()),"idx_business_reviews_user_business_active"){return "",ErrDuplicateBusinessReview};return id,err
}
