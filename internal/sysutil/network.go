package sysutil

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	gnet "github.com/shirou/gopsutil/v3/net"
)

var publicIPClient = &http.Client{Timeout: 2 * time.Second}

func DetectLocalIPv4() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "—"
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP == nil {
				continue
			}

			ip := ipNet.IP.To4()
			if ip == nil {
				continue
			}

			return ip.String()
		}
	}

	return "—"
}

func DetectPublicIP() string {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.ipify.org", nil)
	if err != nil {
		return "—"
	}

	resp, err := publicIPClient.Do(req)
	if err != nil {
		return "—"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64))
	if err != nil {
		return "—"
	}

	ip := strings.TrimSpace(string(body))
	if ip == "" {
		return "—"
	}

	return ip
}

func MeasureTCPPing(address string) float64 {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, 1500*time.Millisecond)
	if err != nil {
		return 0
	}
	_ = conn.Close()

	return float64(time.Since(start).Microseconds()) / 1000.0
}

func IsTCPPortOpen(address string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

type NetworkSampler struct {
	mu           sync.Mutex
	lastSampleAt time.Time
	lastRxBytes  uint64
	lastTxBytes  uint64
}

func NewNetworkSampler() *NetworkSampler {
	return &NetworkSampler{}
}

func (s *NetworkSampler) Sample() (rxTotal, txTotal uint64, rxSpeed, txSpeed float64) {
	stats, err := gnet.IOCounters(false)
	if err != nil || len(stats) == 0 {
		return 0, 0, 0, 0
	}

	rx := stats[0].BytesRecv
	tx := stats[0].BytesSent
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lastSampleAt.IsZero() {
		s.lastSampleAt = now
		s.lastRxBytes = rx
		s.lastTxBytes = tx
		return rx, tx, 0, 0
	}

	elapsed := now.Sub(s.lastSampleAt).Seconds()
	if elapsed <= 0 {
		return rx, tx, 0, 0
	}

	rxSpeed = float64(rx-s.lastRxBytes) / elapsed
	txSpeed = float64(tx-s.lastTxBytes) / elapsed

	s.lastSampleAt = now
	s.lastRxBytes = rx
	s.lastTxBytes = tx

	return rx, tx, rxSpeed, txSpeed
}
