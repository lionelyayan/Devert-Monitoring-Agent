package docker

import (
	"context"
	"fmt"
	"strings"

	networktypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// NetworkInfo holds metadata for a Docker network (Module 6).
type NetworkInfo struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	Driver             string              `json:"driver"`
	Scope              string              `json:"scope"`
	Gateway            string              `json:"gateway"`
	Subnet             string              `json:"subnet"`
	ConnectedContainers []ConnectedContainer `json:"connected_containers"`
}

// ConnectedContainer represents a container connected to a network.
type ConnectedContainer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	IPv4 string `json:"ipv4"`
}

// ListNetworks returns all Docker networks.
func ListNetworks(ctx context.Context, cli *client.Client) ([]NetworkInfo, error) {
	networks, err := cli.NetworkList(ctx, networktypes.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("docker networks: list: %w", err)
	}

	infos := make([]NetworkInfo, 0, len(networks))
	for _, n := range networks {
		gateway := ""
		subnet := ""
		if n.IPAM.Config != nil && len(n.IPAM.Config) > 0 {
			gateway = n.IPAM.Config[0].Gateway
			subnet = n.IPAM.Config[0].Subnet
		}

		var connected []ConnectedContainer
		for id, ep := range n.Containers {
			shortID := id
			if len(id) > 12 {
				shortID = id[:12]
			}
			connected = append(connected, ConnectedContainer{
				ID:   shortID,
				Name: strings.TrimPrefix(ep.Name, "/"),
				IPv4: ep.IPv4Address,
			})
		}

		shortNetID := n.ID
		if len(shortNetID) > 12 {
			shortNetID = shortNetID[:12]
		}

		infos = append(infos, NetworkInfo{
			ID:                  shortNetID,
			Name:                n.Name,
			Driver:              n.Driver,
			Scope:               n.Scope,
			Gateway:             gateway,
			Subnet:              subnet,
			ConnectedContainers: connected,
		})
	}
	return infos, nil
}
