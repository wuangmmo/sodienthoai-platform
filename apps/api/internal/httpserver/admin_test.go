package httpserver

import("net/http";"net/http/httptest";"testing")

func TestAdminUnauthorized(t *testing.T){h:=AdminHandler{Token:"secret"};r:=httptest.NewRequest(http.MethodGet,"/v1/admin/dashboard",nil);w:=httptest.NewRecorder();h.Dashboard(w,r);if w.Code!=http.StatusUnauthorized{t.Fatalf("expected 401 got %d",w.Code)}}
func TestAdminEmptyTokenDenied(t *testing.T){h:=AdminHandler{};r:=httptest.NewRequest(http.MethodGet,"/v1/admin/dashboard",nil);r.Header.Set("Authorization","Bearer ");w:=httptest.NewRecorder();h.Dashboard(w,r);if w.Code!=http.StatusUnauthorized{t.Fatalf("expected 401 got %d",w.Code)}}
func TestAdminAuthorizedConstantTimePath(t *testing.T){h:=AdminHandler{Token:"secret"};r:=httptest.NewRequest(http.MethodGet,"/",nil);r.Header.Set("Authorization","Bearer secret");if !h.authorized(r){t.Fatal("expected valid bearer token to authorize")};r.Header.Set("Authorization","Bearer secrex");if h.authorized(r){t.Fatal("expected invalid bearer token to be denied")}}
