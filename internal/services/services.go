package services

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ServiceStatus represents the health status of a system service.
type ServiceStatus struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Active    bool      `json:"active"`
	CheckedAt time.Time `json:"checked_at"`
}

// DefaultServices lists the services monitored by default.
var DefaultServices = []string{
	"docker",
	"nginx",
	"apache2",
	"php8.1-fpm",
	"php8.2-fpm",
	"php8.3-fpm",
	"postgresql",
	"mysql",
	"redis",
	"rabbitmq-server",
	"n8n",
	"ssh",
	"cron",
}

// CheckAll returns the status of all default services.
func CheckAll() []ServiceStatus {
	results := make([]ServiceStatus, 0, len(DefaultServices))
	for _, name := range DefaultServices {
		results = append(results, Check(name))
	}
	return results
}

// Check queries the status of a single service by name using systemctl.
func Check(name string) ServiceStatus {
	status := querySystemctl(name)
	return ServiceStatus{
		Name:      name,
		Status:    status,
		Active:    status == "running",
		CheckedAt: time.Now(),
	}
}

// querySystemctl calls `systemctl is-active <name>` and interprets the output.
// Possible returns: running, stopped, failed, restarting, unknown
func querySystemctl(name string) string {
	cmd := exec.Command("systemctl", "is-active", name)
	out, err := cmd.Output()
	if err != nil {
		// systemctl exits non-zero for inactive/failed services
		output := strings.TrimSpace(string(out))
		return normalizeStatus(output)
	}
	return normalizeStatus(strings.TrimSpace(string(out)))
}

func normalizeStatus(raw string) string {
	switch raw {
	case "active":
		return "running"
	case "inactive":
		return "stopped"
	case "failed":
		return "failed"
	case "activating", "reloading":
		return "restarting"
	case "":
		return fmt.Sprintf("unknown")
	default:
		return raw
	}
}
