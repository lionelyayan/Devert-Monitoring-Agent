package handlers

import (
	"net/http"

	"github.com/docker/docker/client"

	"github.com/devert/monitor-agent/internal/docker"
)

// ResourcesHandler handles GET /api/resources
type ResourcesHandler struct {
	cli *client.Client
}

func NewResourcesHandler(cli *client.Client) *ResourcesHandler {
	return &ResourcesHandler{cli: cli}
}

func (h *ResourcesHandler) ListResources(w http.ResponseWriter, r *http.Request) {
	resources, err := docker.ListResources(r.Context(), h.cli)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count":     len(resources),
		"resources": resources,
	})
}
