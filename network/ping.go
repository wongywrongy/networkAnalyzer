package network

import (
	"math"
	"time"

	"github.com/go-ping/ping"
)

// PingResult captures a single ping outcome with basic stats.
type PingResult struct {
	Target       string
	PacketsSent  int
	PacketsRecv  int
	PacketLoss   float64
	MinRTT       time.Duration
	MaxRTT       time.Duration
	AvgRTT       time.Duration
	StdDevRTT    time.Duration
	Duration     time.Duration
	ErrorMessage string
}

// PingOnce performs a short ping run (count probes) to a target host/IP.
func PingOnce(target string, count int, timeout time.Duration) PingResult {
	pinger, err := ping.NewPinger(target)
	if err != nil {
		return PingResult{Target: target, ErrorMessage: err.Error()}
	}
	pinger.Count = count
	pinger.Timeout = timeout
	pinger.SetPrivileged(true) // requires admin on Windows

	start := time.Now()
	err = pinger.Run() // blocks until finished
	result := pinger.Statistics()

	res := PingResult{
		Target:      target,
		PacketsSent: result.PacketsSent,
		PacketsRecv: result.PacketsRecv,
		PacketLoss:  result.PacketLoss,
		MinRTT:      result.MinRtt,
		MaxRTT:      result.MaxRtt,
		AvgRTT:      result.AvgRtt,
		StdDevRTT:   result.StdDevRtt,
		Duration:    time.Since(start),
	}
	if err != nil {
		res.ErrorMessage = err.Error()
	}

	// Guard against NaN/Inf values when no replies or stats unavailable.
	res.PacketLoss = sanitizeFloat(res.PacketLoss)
	if res.PacketsSent == 0 && res.PacketsRecv == 0 && res.PacketLoss == 0 {
		res.PacketLoss = 100
	}

	// If we sent packets but received none, set packet loss to 100%
	if res.PacketsSent > 0 && res.PacketsRecv == 0 {
		res.PacketLoss = 100
	}

	// Validate RTT values - if no packets received, durations should be zero/invalid
	if res.PacketsRecv == 0 {
		res.MinRTT = 0
		res.MaxRTT = 0
		res.AvgRTT = 0
		res.StdDevRTT = 0
	}

	return res
}

func sanitizeFloat(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}
