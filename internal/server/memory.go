package server

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/mem"
)

// MemoryInfo holds memory monitoring data.
type MemoryInfo struct {
	TotalBytes     uint64  `json:"total_bytes"`
	UsedBytes      uint64  `json:"used_bytes"`
	FreeBytes      uint64  `json:"free_bytes"`
	CachedBytes    uint64  `json:"cached_bytes"`
	AvailableBytes uint64  `json:"available_bytes"`
	UsedPercent    float64 `json:"used_percent"`
	SwapTotalBytes uint64  `json:"swap_total_bytes"`
	SwapUsedBytes  uint64  `json:"swap_used_bytes"`
	SwapFreeBytes  uint64  `json:"swap_free_bytes"`
	SwapPercent    float64 `json:"swap_percent"`
}

// GetMemory returns current memory and swap metrics.
func GetMemory() (*MemoryInfo, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("server memory: virtual: %w", err)
	}

	swap, err := mem.SwapMemory()
	if err != nil {
		// Swap may not be available on all systems
		swap = &mem.SwapMemoryStat{}
	}

	return &MemoryInfo{
		TotalBytes:     vm.Total,
		UsedBytes:      vm.Used,
		FreeBytes:      vm.Free,
		CachedBytes:    vm.Cached,
		AvailableBytes: vm.Available,
		UsedPercent:    vm.UsedPercent,
		SwapTotalBytes: swap.Total,
		SwapUsedBytes:  swap.Used,
		SwapFreeBytes:  swap.Free,
		SwapPercent:    swap.UsedPercent,
	}, nil
}
