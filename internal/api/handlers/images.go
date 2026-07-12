package handlers

import (
	"net/http"

	"github.com/docker/docker/client"

	"github.com/devert/monitor-agent/internal/docker"
)

// ImagesHandler handles GET /api/images
type ImagesHandler struct {
	cli *client.Client
}

func NewImagesHandler(cli *client.Client) *ImagesHandler {
	return &ImagesHandler{cli: cli}
}

func (h *ImagesHandler) ListImages(w http.ResponseWriter, r *http.Request) {
	images, err := docker.ListImages(r.Context(), h.cli)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count":  len(images),
		"images": images,
	})
}
