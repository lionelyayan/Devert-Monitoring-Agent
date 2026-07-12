package server

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
)

// CPUInfo holds CPU monitoring data.
type CPUInfo struct {
	UsagePercent  float64   `json:"usage_percent"`
	LoadAvg1      float64   `json:"load_avg_1m"`
	LoadAvg5      float64   `json:"load_avg_5m"`
	LoadAvg15     float64   `json:"load_avg_15m"`
	CoreCount     int       `json:"core_count"`
	ThreadCount   int       `json:"thread_count"`
	FrequencyMHz  float64   `json:"frequency_mhz"`
	PerCoreUsage  []float64 `json:"per_core_usage"`
}

// GetCPU returns current CPU metrics.
func GetCPU() (*CPUInfo, error) {
	// Overall CPU usage (non-blocking: interval=0 uses since last call)
	percents, err := cpu.Percent(0, false)
	if err != nil {
		return nil, fmt.Errorf("server cpu: usage percent: %w", err)
	}

	perCore, err := cpu.Percent(0, true)
	if err != nil {
		perCore = []float64{}
	}

	// Physical cores
	physicalCores, _ := cpu.Counts(false)
	// Logical cores (threads)
	logicalCores, _ := cpu.Counts(true)

	// Load average
	avg, err := load.Avg()
	if err != nil {
		avg = &load.AvgStat{}
	}

	// CPU frequency
	var freqMHz float64
	if infos, err := cpu.Info(); err == nil && len(infos) > 0 {
		freqMHz = infos[0].Mhz
	}

	usagePercent := 0.0
	if len(percents) > 0 {
		usagePercent = percents[0]
	}

	return &CPUInfo{
		UsagePercent: usagePercent,
		LoadAvg1:     avg.Load1,
		LoadAvg5:     avg.Load5,
		LoadAvg15:    avg.Load15,
		CoreCount:    physicalCores,
		ThreadCount:  logicalCores,
		FrequencyMHz: freqMHz,
		PerCoreUsage: perCore,
	}, nil
}
