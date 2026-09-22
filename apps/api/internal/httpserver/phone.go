package httpserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/wuangmmo/sodienthoai-platform/apps/api/internal/phone"
)

type PhoneHandler struct{ Repository phone.Repository }

func (h PhoneHandler) Get(w http.ResponseWriter, r *http.Request) {
	e164, err := phone.Normalize(r.PathValue("number"))
	if err != nil {
		writeJSON(w,http.StatusBadRequest,map[string]string{"error":"invalid_phone_number"})
		return
	}

	ctx,cancel := context.WithTimeout(r.Context(),2*time.Second)
	defer cancel()
	result,err := h.Repository.FindByE164(ctx,e164)
	if errors.Is(err,phone.ErrNotFound) {
		writeJSON(w,http.StatusNotFound,map[string]any{
			"error":"phone_number_not_found",
			"number":e164,
			"identified":false,
		})
		return
	}
	if err != nil {
		writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"internal_error"})
		return
	}
	writeJSON(w,http.StatusOK,map[string]any{"data":result})
}
