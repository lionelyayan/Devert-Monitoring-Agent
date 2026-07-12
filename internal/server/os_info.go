package server

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/host"
)

// OSInfo holds operating system information.
type OSInfo struct {
	Hostname      string    `json:"hostname"`
	OS            string    `json:"os"`
	Platform      string    `json:"platform"`
	PlatformVersion string  `json:"platform_version"`
	KernelVersion string    `json:"kernel_version"`
	KernelArch    string    `json:"kernel_arch"`
	UptimeSeconds uint64    `json:"uptime_seconds"`
	BootTime      time.Time `json:"boot_time"`
	Timezone      string    `json:"timezone"`
}

// GetOSInfo returns operating system metadata.
func GetOSInfo() (*OSInfo, error) {
	info, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("server os: host info: %w", err)
	}

	bootTime := time.Unix(int64(info.BootTime), 0)
	tz, _ := time.Now().Zone()

	return &OSInfo{
		Hostname:        info.Hostname,
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		KernelArch:      info.KernelArch,
		UptimeSeconds:   info.Uptime,
		BootTime:        bootTime,
		Timezone:        tz,
	}, nil
}
