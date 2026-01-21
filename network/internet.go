package network

import (
	"context"
	"net"
	"net/http"
	"time"
)

// InternetProbeResult captures simple connectivity metrics for the local device.
type InternetProbeResult struct {
	ResolvedHosts  map[string]bool    `json:"resolvedHosts"`
	ResolveTimeMs  float64            `json:"resolveTimeMs"`
	ConnectTimes   map[string]float64 `json:"connectTimesMs"`
	HTTPStatus     int                `json:"httpStatus"`
	HTTPDurationMs float64            `json:"httpDurationMs"`
	ErrorMessage   string             `json:"errorMessage,omitempty"`
}

// ProbeInternet performs lightweight DNS, TCP, and HTTP reachability checks.
func ProbeInternet(ctx context.Context) InternetProbeResult {
	result := InternetProbeResult{
		ResolvedHosts: make(map[string]bool),
		ConnectTimes:  make(map[string]float64),
	}

	resolveStart := time.Now()
	hosts := []string{"www.google.com"}
	for _, h := range hosts {
		addrs, err := net.DefaultResolver.LookupHost(ctx, h)
		if err == nil && len(addrs) > 0 {
			result.ResolvedHosts[h] = true
		} else {
			result.ResolvedHosts[h] = false
		}
	}
	result.ResolveTimeMs = float64(time.Since(resolveStart).Milliseconds())

	dialTargets := []string{"1.1.1.1:443", "8.8.8.8:443"}
	for _, t := range dialTargets {
		start := time.Now()
		dialer := &net.Dialer{Timeout: 3 * time.Second}
		conn, err := dialer.DialContext(ctx, "tcp", t)
		if err == nil {
			result.ConnectTimes[t] = float64(time.Since(start).Milliseconds())
			conn.Close()
		} else {
			result.ConnectTimes[t] = -1
		}
	}

	// Simple HTTP check to a 204 endpoint
	httpClient := &http.Client{Timeout: 4 * time.Second}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.google.com/generate_204", nil)
	start := time.Now()
	resp, err := httpClient.Do(req)
	if err != nil {
		result.HTTPStatus = 0
		result.HTTPDurationMs = -1
		result.ErrorMessage = err.Error()
		return result
	}
	defer resp.Body.Close()
	result.HTTPStatus = resp.StatusCode
	result.HTTPDurationMs = float64(time.Since(start).Milliseconds())

	return result
}
