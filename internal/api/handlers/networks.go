package handlers

import (
	"net/http"

	"github.com/docker/docker/client"

	"github.com/devert/monitor-agent/internal/docker"
)

// NetworksHandler handles GET /api/networks
type NetworksHandler struct {
	cli *client.Client
}

func NewNetworksHandler(cli *client.Client) *NetworksHandler {
	return &NetworksHandler{cli: cli}
}

func (h *NetworksHandler) ListNetworks(w http.ResponseWriter, r *http.Request) {
	networks, err := docker.ListNetworks(r.Context(), h.cli)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count":    len(networks),
		"networks": networks,
	})
}
