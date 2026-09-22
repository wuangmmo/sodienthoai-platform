package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitByIP(t *testing.T) {
	h := RateLimitByIP(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, 2, time.Minute)

	for i, want := range []int{http.StatusNoContent, http.StatusNoContent, http.StatusTooManyRequests} {
		req := httptest.NewRequest(http.MethodGet, "/v1/phone/%2B84705899899", nil)
		req.RemoteAddr = "203.0.113.10:12345"
		rec := httptest.NewRecorder()
		h(rec, req)
		if rec.Code != want {
			t.Fatalf("request %d: expected %d, got %d", i+1, want, rec.Code)
		}
	}
}
