package handlers

import (
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/go-chi/chi/v5"

	"github.com/devert/monitor-agent/internal/docker"
)

// ContainersHandler handles container-related API endpoints.
type ContainersHandler struct {
	cli *client.Client
}

func NewContainersHandler(cli *client.Client) *ContainersHandler {
	return &ContainersHandler{cli: cli}
}

// ListContainers handles GET /api/containers
func (h *ContainersHandler) ListContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := docker.ListContainers(r.Context(), h.cli)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count":      len(containers),
		"containers": containers,
	})
}

// GetContainer handles GET /api/container/{id}
func (h *ContainersHandler) GetContainer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	info, err := docker.InspectContainer(r.Context(), h.cli, id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, info)
}

// StartContainer handles POST /api/container/{id}/start
func (h *ContainersHandler) StartContainer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.cli.ContainerStart(r.Context(), id, container.StartOptions{}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "started", "container": id})
}

// StopContainer handles POST /api/container/{id}/stop
func (h *ContainersHandler) StopContainer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	timeout := 10
	if err := h.cli.ContainerStop(r.Context(), id, container.StopOptions{Timeout: &timeout}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped", "container": id})
}

// RestartContainer handles POST /api/container/{id}/restart
func (h *ContainersHandler) RestartContainer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	timeout := 10
	if err := h.cli.ContainerRestart(r.Context(), id, container.StopOptions{Timeout: &timeout}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "restarted", "container": id})
}

// RemoveContainer handles POST /api/container/{id}/remove
func (h *ContainersHandler) RemoveContainer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.cli.ContainerRemove(r.Context(), id, container.RemoveOptions{Force: true}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed", "container": id})
}
