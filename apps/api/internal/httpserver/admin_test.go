package httpserver

import("net/http";"net/http/httptest";"testing")

func TestAdminUnauthorized(t *testing.T){h:=AdminHandler{Token:"secret"};r:=httptest.NewRequest(http.MethodGet,"/v1/admin/dashboard",nil);w:=httptest.NewRecorder();h.Dashboard(w,r);if w.Code!=http.StatusUnauthorized{t.Fatalf("expected 401 got %d",w.Code)}}
func TestAdminEmptyTokenDenied(t *testing.T){h:=AdminHandler{};r:=httptest.NewRequest(http.MethodGet,"/v1/admin/dashboard",nil);r.Header.Set("Authorization","Bearer ");w:=httptest.NewRecorder();h.Dashboard(w,r);if w.Code!=http.StatusUnauthorized{t.Fatalf("expected 401 got %d",w.Code)}}
func TestAdminAuthorizedConstantTimePath(t *testing.T){h:=AdminHandler{Token:"secret"};r:=httptest.NewRequest(http.MethodGet,"/",nil);r.Header.Set("Authorization","Bearer secret");if !h.authorized(r){t.Fatal("expected valid bearer token to authorize")};r.Header.Set("Authorization","Bearer secrex");if h.authorized(r){t.Fatal("expected invalid bearer token to be denied")}}

func TestNewAdminQueuesUnauthorized(t *testing.T){
 h:=AdminHandler{Token:"secret"}
 cases:=[]struct{name string;fn func(http.ResponseWriter,*http.Request)}{{"appeals",h.Appeals},{"business_verifications",h.BusinessVerifications},{"business_reviews",h.BusinessReviews}}
 for _,tc:=range cases{t.Run(tc.name,func(t *testing.T){r:=httptest.NewRequest(http.MethodGet,"/",nil);w:=httptest.NewRecorder();tc.fn(w,r);if w.Code!=http.StatusUnauthorized{t.Fatalf("expected 401 got %d",w.Code)}})}
}
func TestNewModerationUnauthorized(t *testing.T){
 h:=ModerationHandler{Token:"secret"}
 cases:=[]struct{name string;fn func(http.ResponseWriter,*http.Request)}{{"appeal",h.Appeal},{"business_verification",h.BusinessVerification},{"business_review",h.BusinessReview}}
 for _,tc:=range cases{t.Run(tc.name,func(t *testing.T){r:=httptest.NewRequest(http.MethodPatch,"/",nil);w:=httptest.NewRecorder();tc.fn(w,r);if w.Code!=http.StatusUnauthorized{t.Fatalf("expected 401 got %d",w.Code)}})}
}
