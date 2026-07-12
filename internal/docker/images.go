package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// ImageInfo holds metadata for a Docker image (Module 4).
type ImageInfo struct {
	Repository   string    `json:"repository"`
	Tag          string    `json:"tag"`
	ImageID      string    `json:"image_id"`
	Size         int64     `json:"size_bytes"`
	CreatedAt    time.Time `json:"created_at"`
	Dangling     bool      `json:"dangling"`
	UsedBy       []string  `json:"used_by_containers"`
}

// ListImages returns all Docker images with metadata.
func ListImages(ctx context.Context, cli *client.Client) ([]ImageInfo, error) {
	images, err := cli.ImageList(ctx, image.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("docker images: list: %w", err)
	}

	// Build a map of image ID -> containers using them
	usageMap, err := buildImageUsageMap(ctx, cli)
	if err != nil {
		usageMap = map[string][]string{}
	}

	infos := make([]ImageInfo, 0, len(images))
	for _, img := range images {
		repo := "<none>"
		tag := "<none>"
		dangling := true

		if len(img.RepoTags) > 0 {
			dangling = false
			parts := strings.SplitN(img.RepoTags[0], ":", 2)
			repo = parts[0]
			if len(parts) == 2 {
				tag = parts[1]
			}
		}

		shortID := img.ID
		if strings.HasPrefix(shortID, "sha256:") {
			shortID = shortID[7:19]
		}

		infos = append(infos, ImageInfo{
			Repository: repo,
			Tag:        tag,
			ImageID:    shortID,
			Size:       img.Size,
			CreatedAt:  time.Unix(img.Created, 0),
			Dangling:   dangling,
			UsedBy:     usageMap[img.ID],
		})
	}
	return infos, nil
}

// buildImageUsageMap maps image IDs to container names using them.
func buildImageUsageMap(ctx context.Context, cli *client.Client) (map[string][]string, error) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("docker images: list containers for usage map: %w", err)
	}

	m := make(map[string][]string)
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		m[c.ImageID] = append(m[c.ImageID], name)
	}
	return m, nil
}
