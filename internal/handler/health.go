package handler

import (
	"context"
	"encoding/json"
	"net/http"

	api "github.com/freyliee/template/internal/generated"
)

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(api.HealthResponse{
		Status: api.Ok,
	})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusOK
	response := api.HealthResponse{
		Status: api.Ok,
	}

	timeoutCtx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	if err := h.pool.Ping(timeoutCtx); err != nil {
		statusCode = http.StatusServiceUnavailable
		response.Status = api.Unavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(response)
}
