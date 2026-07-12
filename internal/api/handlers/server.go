package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/devert/monitor-agent/internal/server"
)

// ServerHandler handles GET /api/server
type ServerHandler struct {
	serverName string
}

func NewServerHandler(serverName string) *ServerHandler {
	return &ServerHandler{serverName: serverName}
}

// GetServer returns a comprehensive server snapshot.
func (h *ServerHandler) GetServer(w http.ResponseWriter, r *http.Request) {
	type ServerSnapshot struct {
		Name   string          `json:"server_name"`
		CPU    *server.CPUInfo    `json:"cpu"`
		Memory *server.MemoryInfo `json:"memory"`
		Disk   *server.DiskInfo   `json:"disk"`
		Network *server.NetworkInfo `json:"network"`
		OS     *server.OSInfo     `json:"os"`
		Errors []string          `json:"errors,omitempty"`
	}

	snap := ServerSnapshot{Name: h.serverName}
	var errs []string

	cpu, err := server.GetCPU()
	if err != nil {
		errs = append(errs, "cpu: "+err.Error())
	} else {
		snap.CPU = cpu
	}

	mem, err := server.GetMemory()
	if err != nil {
		errs = append(errs, "memory: "+err.Error())
	} else {
		snap.Memory = mem
	}

	disk, err := server.GetDisk()
	if err != nil {
		errs = append(errs, "disk: "+err.Error())
	} else {
		snap.Disk = disk
	}

	net, err := server.GetNetwork()
	if err != nil {
		errs = append(errs, "network: "+err.Error())
	} else {
		snap.Network = net
	}

	osInfo, err := server.GetOSInfo()
	if err != nil {
		errs = append(errs, "os: "+err.Error())
	} else {
		snap.OS = osInfo
	}

	if len(errs) > 0 {
		snap.Errors = errs
	}

	writeJSON(w, http.StatusOK, snap)
}

// writeJSON is a shared helper that writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
