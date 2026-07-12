package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// VolumeInfo holds metadata for a Docker volume (Module 5).
type VolumeInfo struct {
	Name       string    `json:"name"`
	MountPoint string    `json:"mount_point"`
	Driver     string    `json:"driver"`
	CreatedAt  time.Time `json:"created_at"`
	UsedBy     []string  `json:"used_by_containers"`
}

// ListVolumes returns all Docker volumes.
func ListVolumes(ctx context.Context, cli *client.Client) ([]VolumeInfo, error) {
	resp, err := cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("docker volumes: list: %w", err)
	}

	usageMap, err := buildVolumeUsageMap(ctx, cli)
	if err != nil {
		usageMap = map[string][]string{}
	}

	infos := make([]VolumeInfo, 0, len(resp.Volumes))
	for _, v := range resp.Volumes {
		createdAt := time.Time{}
		if v.CreatedAt != "" {
			createdAt, _ = time.Parse(time.RFC3339, v.CreatedAt)
		}

		infos = append(infos, VolumeInfo{
			Name:       v.Name,
			MountPoint: v.Mountpoint,
			Driver:     v.Driver,
			CreatedAt:  createdAt,
			UsedBy:     usageMap[v.Name],
		})
	}
	return infos, nil
}

// buildVolumeUsageMap maps volume names to container names using them.
func buildVolumeUsageMap(ctx context.Context, cli *client.Client) (map[string][]string, error) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("docker volumes: list containers for usage map: %w", err)
	}

	m := make(map[string][]string)
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		for _, mount := range c.Mounts {
			if mount.Type == "volume" {
				m[mount.Name] = append(m[mount.Name], name)
			}
		}
	}
	return m, nil
}
