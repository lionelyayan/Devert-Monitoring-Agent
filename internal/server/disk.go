package server

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/disk"
)

// DiskPartition holds usage stats for a single disk partition.
type DiskPartition struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	TotalBytes  uint64  `json:"total_bytes"`
	UsedBytes   uint64  `json:"used_bytes"`
	FreeBytes   uint64  `json:"free_bytes"`
	UsedPercent float64 `json:"used_percent"`
}

// DiskIOStats holds disk I/O counters.
type DiskIOStats struct {
	Device     string `json:"device"`
	ReadBytes  uint64 `json:"read_bytes"`
	WriteBytes uint64 `json:"write_bytes"`
	ReadCount  uint64 `json:"read_count"`
	WriteCount uint64 `json:"write_count"`
}

// DiskInfo holds all disk monitoring data.
type DiskInfo struct {
	Partitions []DiskPartition `json:"partitions"`
	IO         []DiskIOStats   `json:"io"`
}

// GetDisk returns disk usage for all partitions and I/O counters.
func GetDisk() (*DiskInfo, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("server disk: partitions: %w", err)
	}

	var partitions []DiskPartition
	for _, p := range parts {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		partitions = append(partitions, DiskPartition{
			Device:      p.Device,
			Mountpoint:  p.Mountpoint,
			Fstype:      p.Fstype,
			TotalBytes:  usage.Total,
			UsedBytes:   usage.Used,
			FreeBytes:   usage.Free,
			UsedPercent: usage.UsedPercent,
		})
	}

	// I/O counters (best-effort; may be empty on some systems)
	ioCounters, err := disk.IOCounters()
	var ios []DiskIOStats
	if err == nil {
		for device, c := range ioCounters {
			ios = append(ios, DiskIOStats{
				Device:     device,
				ReadBytes:  c.ReadBytes,
				WriteBytes: c.WriteBytes,
				ReadCount:  c.ReadCount,
				WriteCount: c.WriteCount,
			})
		}
	}

	return &DiskInfo{
		Partitions: partitions,
		IO:         ios,
	}, nil
}
