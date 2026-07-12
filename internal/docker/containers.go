package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// ContainerInfo holds full metadata for a single container (Module 2).
type ContainerInfo struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Image        string            `json:"image"`
	ImageVersion string            `json:"image_version"`
	Status       string            `json:"status"`
	State        string            `json:"state"`
	RestartCount int               `json:"restart_count"`
	ExitCode     int               `json:"exit_code"`
	CreatedAt    time.Time         `json:"created_at"`
	StartedAt    string            `json:"started_at"`
	FinishedAt   string            `json:"finished_at"`
	UptimeSeconds int64            `json:"uptime_seconds"`
	Health       string            `json:"health"`
	Labels       map[string]string `json:"labels"`
	Env          []string          `json:"env"`
	Networks     []string          `json:"networks"`
	Mounts       []MountInfo       `json:"mounts"`
	Ports        []PortInfo        `json:"ports"`
}

// MountInfo represents a mounted volume.
type MountInfo struct {
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
	RW          bool   `json:"rw"`
}

// PortInfo represents a published port mapping.
type PortInfo struct {
	PrivatePort uint16 `json:"private_port"`
	PublicPort  uint16 `json:"public_port"`
	Type        string `json:"type"`
	IP          string `json:"ip"`
}

// ListContainers returns full metadata for all containers (running and stopped).
func ListContainers(ctx context.Context, cli *client.Client) ([]ContainerInfo, error) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("docker containers: list: %w", err)
	}

	infos := make([]ContainerInfo, 0, len(containers))
	for _, c := range containers {
		info, err := InspectContainer(ctx, cli, c.ID)
		if err != nil {
			continue
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// InspectContainer returns detailed metadata for a single container by ID or name.
func InspectContainer(ctx context.Context, cli *client.Client, idOrName string) (ContainerInfo, error) {
	data, err := cli.ContainerInspect(ctx, idOrName)
	if err != nil {
		return ContainerInfo{}, fmt.Errorf("docker containers: inspect %s: %w", idOrName, err)
	}

	return buildContainerInfo(data), nil
}

func buildContainerInfo(data types.ContainerJSON) ContainerInfo {
	name := data.Name
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}

	createdAt, _ := time.Parse(time.RFC3339Nano, data.Created)

	// Calculate uptime
	var uptimeSeconds int64
	if data.State != nil && data.State.Running && data.State.StartedAt != "" {
		if startedAt, err := time.Parse(time.RFC3339Nano, data.State.StartedAt); err == nil {
			uptimeSeconds = int64(time.Since(startedAt).Seconds())
		}
	}

	// Health status
	health := "none"
	if data.State != nil && data.State.Health != nil {
		health = data.State.Health.Status
	}

	// Networks
	var networks []string
	if data.NetworkSettings != nil {
		for netName := range data.NetworkSettings.Networks {
			networks = append(networks, netName)
		}
	}

	// Mounts
	var mounts []MountInfo
	for _, m := range data.Mounts {
		mounts = append(mounts, MountInfo{
			Type:        string(m.Type),
			Source:      m.Source,
			Destination: m.Destination,
			Mode:        m.Mode,
			RW:          m.RW,
		})
	}

	// Ports
	var ports []PortInfo
	if data.NetworkSettings != nil {
		for privatePort, bindings := range data.NetworkSettings.Ports {
			for _, b := range bindings {
				var pubPort uint16
				if b.HostPort != "" {
					fmt.Sscanf(b.HostPort, "%d", &pubPort)
				}
				ports = append(ports, PortInfo{
					PrivatePort: uint16(privatePort.Int()),
					PublicPort:  pubPort,
					Type:        privatePort.Proto(),
					IP:          b.HostIP,
				})
			}
		}
	}

	exitCode := 0
	startedAt := ""
	finishedAt := ""
	state := "unknown"
	restartCount := 0

	if data.State != nil {
		exitCode = data.State.ExitCode
		startedAt = data.State.StartedAt
		finishedAt = data.State.FinishedAt
		state = data.State.Status
	}
	if data.RestartCount != 0 {
		restartCount = data.RestartCount
	}

	// Image version from RepoTags
	image := data.Config.Image
	imageVersion := "latest"
	if data.Config != nil && len(data.ContainerJSONBase.Image) > 0 {
		imageVersion = data.Config.Image
	}

	env := []string{}
	if data.Config != nil {
		env = data.Config.Env
	}

	return ContainerInfo{
		ID:            data.ID[:12],
		Name:          name,
		Image:         image,
		ImageVersion:  imageVersion,
		Status:        data.State.Status,
		State:         state,
		RestartCount:  restartCount,
		ExitCode:      exitCode,
		CreatedAt:     createdAt,
		StartedAt:     startedAt,
		FinishedAt:    finishedAt,
		UptimeSeconds: uptimeSeconds,
		Health:        health,
		Labels:        data.Config.Labels,
		Env:           env,
		Networks:      networks,
		Mounts:        mounts,
		Ports:         ports,
	}
}
