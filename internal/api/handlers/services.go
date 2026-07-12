package handlers

import (
	"net/http"

	"github.com/devert/monitor-agent/internal/services"
)

// ServicesHandler handles GET /api/services
type ServicesHandler struct{}

func NewServicesHandler() *ServicesHandler {
	return &ServicesHandler{}
}

func (h *ServicesHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	statuses := services.CheckAll()
	writeJSON(w, http.StatusOK, map[string]any{
		"count":    len(statuses),
		"services": statuses,
	})
}
