package handlers

import (
	"net/http"

	"github.com/docker/docker/client"

	"github.com/devert/monitor-agent/internal/docker"
)

// VolumesHandler handles GET /api/volumes
type VolumesHandler struct {
	cli *client.Client
}

func NewVolumesHandler(cli *client.Client) *VolumesHandler {
	return &VolumesHandler{cli: cli}
}

func (h *VolumesHandler) ListVolumes(w http.ResponseWriter, r *http.Request) {
	volumes, err := docker.ListVolumes(r.Context(), h.cli)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count":   len(volumes),
		"volumes": volumes,
	})
}
