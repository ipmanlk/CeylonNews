package httpx

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}

func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
		Code:    status,
	})
}

func RespondBadRequest(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusBadRequest, message)
}

func RespondNotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "resource not found"
	}
	RespondError(w, http.StatusNotFound, message)
}

func RespondInternalError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "internal server error"
	}
	RespondError(w, http.StatusInternalServerError, message)
}
