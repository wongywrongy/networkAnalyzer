package network

import (
	"bytes"
	"os/exec"
	"strings"
)

// DiscoveryEntry represents a row from ARP discovery.
type DiscoveryEntry struct {
	IP   string `json:"ip"`
	MAC  string `json:"mac"`
	Type string `json:"type"`
}

// DiscoverLAN executes "arp -a" (Windows-friendly) and parses entries.
func DiscoverLAN() ([]DiscoveryEntry, error) {
	cmd := exec.Command("arp", "-a")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(buf.String(), "\n")
	var entries []DiscoveryEntry
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(strings.ToLower(line), "interface:") || strings.HasPrefix(strings.ToLower(line), "internet") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		ip := fields[0]
		mac := fields[1]
		typ := fields[2]
		entries = append(entries, DiscoveryEntry{IP: ip, MAC: mac, Type: typ})
	}
	return entries, nil
}

