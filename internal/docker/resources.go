package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// ContainerResource holds real-time resource usage for a single container (Module 3).
type ContainerResource struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryUsage   uint64  `json:"memory_usage_bytes"`
	MemoryLimit   uint64  `json:"memory_limit_bytes"`
	MemoryPercent float64 `json:"memory_percent"`
	NetworkRX     uint64  `json:"network_rx_bytes"`
	NetworkTX     uint64  `json:"network_tx_bytes"`
	DiskRead      uint64  `json:"disk_read_bytes"`
	DiskWrite     uint64  `json:"disk_write_bytes"`
	ProcessCount  uint32  `json:"process_count"`
}

// statsJSON is a local struct that mirrors the Docker Stats JSON payload.
// Defined locally to avoid Docker SDK version-specific type changes.
type statsJSON struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	CPUStats struct {
		CPUUsage struct {
			TotalUsage  uint64   `json:"total_usage"`
			PercpuUsage []uint64 `json:"percpu_usage"`
		} `json:"cpu_usage"`
		SystemUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs  uint32 `json:"online_cpus"`
	} `json:"cpu_stats"`

	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemUsage uint64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`

	MemoryStats struct {
		Usage uint64            `json:"usage"`
		Limit uint64            `json:"limit"`
		Stats map[string]uint64 `json:"stats"`
	} `json:"memory_stats"`

	Networks map[string]struct {
		RxBytes uint64 `json:"rx_bytes"`
		TxBytes uint64 `json:"tx_bytes"`
	} `json:"networks"`

	BlkioStats struct {
		IoServiceBytesRecursive []struct {
			Op    string `json:"op"`
			Value uint64 `json:"value"`
		} `json:"io_service_bytes_recursive"`
	} `json:"blkio_stats"`

	PidsStats struct {
		Current uint32 `json:"current"`
	} `json:"pids_stats"`
}

// ListResources returns current resource usage for all running containers.
func ListResources(ctx context.Context, cli *client.Client) ([]ContainerResource, error) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("docker resources: list containers: %w", err)
	}

	resources := make([]ContainerResource, 0, len(containers))
	for _, c := range containers {
		res, err := GetContainerResource(ctx, cli, c.ID)
		if err != nil {
			continue
		}
		resources = append(resources, res)
	}
	return resources, nil
}

// GetContainerResource returns resource usage for a single container.
func GetContainerResource(ctx context.Context, cli *client.Client, idOrName string) (ContainerResource, error) {
	resp, err := cli.ContainerStats(ctx, idOrName, false)
	if err != nil {
		return ContainerResource{}, fmt.Errorf("docker resources: stats %s: %w", idOrName, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ContainerResource{}, fmt.Errorf("docker resources: read body: %w", err)
	}

	var stats statsJSON
	if err := json.Unmarshal(body, &stats); err != nil {
		return ContainerResource{}, fmt.Errorf("docker resources: parse stats: %w", err)
	}

	name := stats.Name
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}

	cpuPercent := calcCPUPercent(&stats)
	memPercent := 0.0
	if stats.MemoryStats.Limit > 0 {
		memPercent = float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100.0
	}

	var netRX, netTX uint64
	for _, netStats := range stats.Networks {
		netRX += netStats.RxBytes
		netTX += netStats.TxBytes
	}

	var diskRead, diskWrite uint64
	for _, blkio := range stats.BlkioStats.IoServiceBytesRecursive {
		switch blkio.Op {
		case "Read":
			diskRead += blkio.Value
		case "Write":
			diskWrite += blkio.Value
		}
	}

	shortID := idOrName
	if len(shortID) > 12 {
		shortID = shortID[:12]
	}

	return ContainerResource{
		ID:            shortID,
		Name:          name,
		CPUPercent:    cpuPercent,
		MemoryUsage:   stats.MemoryStats.Usage,
		MemoryLimit:   stats.MemoryStats.Limit,
		MemoryPercent: memPercent,
		NetworkRX:     netRX,
		NetworkTX:     netTX,
		DiskRead:      diskRead,
		DiskWrite:     diskWrite,
		ProcessCount:  stats.PidsStats.Current,
	}, nil
}

func calcCPUPercent(s *statsJSON) float64 {
	cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage) - float64(s.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(s.CPUStats.SystemUsage) - float64(s.PreCPUStats.SystemUsage)
	numCPU := float64(s.CPUStats.OnlineCPUs)
	if numCPU == 0 {
		numCPU = float64(len(s.CPUStats.CPUUsage.PercpuUsage))
	}
	if systemDelta > 0 && cpuDelta > 0 {
		return (cpuDelta / systemDelta) * numCPU * 100.0
	}
	return 0.0
}
