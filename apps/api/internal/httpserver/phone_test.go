package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPhoneHandlerRejectsNationalFormatWithoutCountryContext(t *testing.T) {
	h := PhoneHandler{}
	req := httptest.NewRequest(http.MethodGet, "/v1/phone/0705899899", nil)
	req.SetPathValue("number", "0705899899")
	rec := httptest.NewRecorder()

	h.Get(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
