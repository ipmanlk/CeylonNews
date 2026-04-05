package handlers

import (
	"ipmanlk/cnapi/pkg/httpx"
	"net/http"
	"time"
)

func Health(w http.ResponseWriter, r *http.Request) {
	httpx.RespondJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
