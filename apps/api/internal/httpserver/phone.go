package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/phone"
)

type PhoneHandler struct{ Service phone.Service }

func (h PhoneHandler) Get(w http.ResponseWriter, r *http.Request) {
	e164, err := phone.Normalize(r.PathValue("number"))
	if err != nil {
		writeJSON(w,http.StatusBadRequest,map[string]string{"error":"invalid_phone_number"})
		return
	}

	ctx,cancel := context.WithTimeout(r.Context(),2*time.Second)
	defer cancel()
	result,err := h.Service.Find(ctx,e164)
	if phone.IsNotFound(err) {
		writeJSON(w,http.StatusNotFound,map[string]any{"error":"phone_number_not_found","number":e164,"identified":false})
		return
	}
	if err != nil {
		writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"})
		return
	}
	writeJSON(w,http.StatusOK,map[string]any{"data":result})
}
