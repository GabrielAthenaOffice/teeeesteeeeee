package handler

import (
	"encoding/json"
	"net/http"

	"github.com/CunhazadanoDale/trads-market-test/internal/adapter/http/dtos"
)

const (
	CodeInvalidRequest = "invalid_request"
	CodeStateNotFound  = "state_not_found"
	CodeCityNotFound   = "city_not_found"
	CodeNotFound       = "not_found"
	CodeInternalError  = "internal_error"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, dtos.NewErrorResponse(code, message))
}
