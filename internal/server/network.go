package server

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/net"
)

// NetworkInterface holds stats for a single network interface.
type NetworkInterface struct {
	Name       string `json:"name"`
	BytesRecv  uint64 `json:"bytes_recv"`
	BytesSent  uint64 `json:"bytes_sent"`
	PacketRecv uint64 `json:"packets_recv"`
	PacketSent uint64 `json:"packets_sent"`
	ErrIn      uint64 `json:"errors_in"`
	ErrOut     uint64 `json:"errors_out"`
	DropIn     uint64 `json:"drops_in"`
	DropOut    uint64 `json:"drops_out"`
}

// NetworkInfo holds all server network monitoring data.
type NetworkInfo struct {
	LocalIPs   []string           `json:"local_ips"`
	PublicIP   string             `json:"public_ip"`
	Interfaces []NetworkInterface `json:"interfaces"`
}

// GetNetwork returns network interface stats and IP addresses.
func GetNetwork() (*NetworkInfo, error) {
	counters, err := net.IOCounters(true)
	if err != nil {
		return nil, fmt.Errorf("server network: io counters: %w", err)
	}

	var ifaces []NetworkInterface
	for _, c := range counters {
		if c.Name == "lo" {
			continue // Skip loopback
		}
		ifaces = append(ifaces, NetworkInterface{
			Name:       c.Name,
			BytesRecv:  c.BytesRecv,
			BytesSent:  c.BytesSent,
			PacketRecv: c.PacketsRecv,
			PacketSent: c.PacketsSent,
			ErrIn:      c.Errin,
			ErrOut:     c.Errout,
			DropIn:     c.Dropin,
			DropOut:    c.Dropout,
		})
	}

	localIPs := getLocalIPs()
	publicIP := getPublicIP()

	return &NetworkInfo{
		LocalIPs:   localIPs,
		PublicIP:   publicIP,
		Interfaces: ifaces,
	}, nil
}

func getLocalIPs() []string {
	addrs, err := net.Interfaces()
	if err != nil {
		return nil
	}

	var ips []string
	for _, iface := range addrs {
		for _, addr := range iface.Addrs {
			ip := addr.Addr
			// Strip CIDR notation if present
			if idx := strings.Index(ip, "/"); idx != -1 {
				ip = ip[:idx]
			}
			// Skip loopback and IPv6 link-local
			if ip != "127.0.0.1" && !strings.HasPrefix(ip, "::") && !strings.HasPrefix(ip, "fe80") {
				ips = append(ips, ip)
			}
		}
	}
	return ips
}

func getPublicIP() string {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(body))
}
