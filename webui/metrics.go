package webui

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

type MetricsResponse struct {
	OK          bool            `json:"ok"`
	CollectedAt string          `json:"collectedAt,omitempty"`
	Overview    OverviewMetrics `json:"overview"`
	System      SystemMetrics   `json:"system"`
	Network     NetworkMetrics  `json:"network"`
	Error       string          `json:"error,omitempty"`
}

type OverviewMetrics struct {
	CPUUsagePercent       float64 `json:"cpuUsagePercent"`
	CPUTemperatureCelsius float64 `json:"cpuTemperatureCelsius"`
	RAMUsedBytes          uint64  `json:"ramUsedBytes"`
	RAMTotalBytes         uint64  `json:"ramTotalBytes"`
	RAMUsagePercent       float64 `json:"ramUsagePercent"`
	DiskUsedBytes         uint64  `json:"diskUsedBytes"`
	DiskTotalBytes        uint64  `json:"diskTotalBytes"`
	DiskUsagePercent      float64 `json:"diskUsagePercent"`
	UptimeSeconds         uint64  `json:"uptimeSeconds"`
}

type SystemMetrics struct {
	Hostname        string  `json:"hostname"`
	Platform        string  `json:"platform"`
	PlatformVersion string  `json:"platformVersion"`
	KernelVersion   string  `json:"kernelVersion"`
	Architecture    string  `json:"architecture"`
	Load1           float64 `json:"load1"`
	Load5           float64 `json:"load5"`
	Load15          float64 `json:"load15"`
	Processes       uint64  `json:"processes"`
	BootTimeUnix    uint64  `json:"bootTimeUnix"`
	LogicalCPUCount int     `json:"logicalCpuCount"`
	CPUModel        string  `json:"cpuModel"`
	CPUFrequencyMHz float64 `json:"cpuFrequencyMHz"`
}

type NetworkMetrics struct {
	LocalIPv4 string `json:"localIpv4"`
}

func collectMetrics() (MetricsResponse, error) {
	var resp MetricsResponse
	resp.OK = true
	resp.CollectedAt = time.Now().Format(time.RFC3339)

	vm, err := mem.VirtualMemory()
	if err != nil {
		return MetricsResponse{}, fmt.Errorf("virtual memory: %w", err)
	}

	diskUsage, err := disk.Usage("/")
	if err != nil {
		return MetricsResponse{}, fmt.Errorf("disk usage: %w", err)
	}

	hostInfo, err := host.Info()
	if err != nil {
		return MetricsResponse{}, fmt.Errorf("host info: %w", err)
	}

	loadAvg, err := load.Avg()
	if err != nil {
		return MetricsResponse{}, fmt.Errorf("load avg: %w", err)
	}

	cpuUsagePercent := 0.0
	cpuPercents, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercents) > 0 {
		cpuUsagePercent = cpuPercents[0]
	}

	cpuInfos, _ := cpu.Info()
	cpuModel := ""
	cpuFrequencyMHz := 0.0
	if len(cpuInfos) > 0 {
		cpuModel = cpuInfos[0].ModelName
		cpuFrequencyMHz = cpuInfos[0].Mhz
	}

	logicalCPUCount, _ := cpu.Counts(true)

	hostname, _ := os.Hostname()

	resp.Overview = OverviewMetrics{
		CPUUsagePercent:       cpuUsagePercent,
		CPUTemperatureCelsius: readCPUTemperature(),
		RAMUsedBytes:          vm.Used,
		RAMTotalBytes:         vm.Total,
		RAMUsagePercent:       vm.UsedPercent,
		DiskUsedBytes:         diskUsage.Used,
		DiskTotalBytes:        diskUsage.Total,
		DiskUsagePercent:      diskUsage.UsedPercent,
		UptimeSeconds:         hostInfo.Uptime,
	}

	resp.System = SystemMetrics{
		Hostname:        hostname,
		Platform:        hostInfo.Platform,
		PlatformVersion: hostInfo.PlatformVersion,
		KernelVersion:   hostInfo.KernelVersion,
		Architecture:    runtime.GOARCH,
		Load1:           loadAvg.Load1,
		Load5:           loadAvg.Load5,
		Load15:          loadAvg.Load15,
		Processes:       hostInfo.Procs,
		BootTimeUnix:    hostInfo.BootTime,
		LogicalCPUCount: logicalCPUCount,
		CPUModel:        cpuModel,
		CPUFrequencyMHz: cpuFrequencyMHz,
	}

	resp.Network = NetworkMetrics{
		LocalIPv4: detectLocalIPv4(),
	}

	return resp, nil
}

func detectLocalIPv4() string {
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

func readCPUTemperature() float64 {
	raw, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err == nil {
		value, parseErr := strconv.ParseFloat(strings.TrimSpace(string(raw)), 64)
		if parseErr == nil {
			return value / 1000.0
		}
	}

	temps, err := host.SensorsTemperatures()
	if err == nil {
		for _, t := range temps {
			if t.Temperature > 0 {
				return t.Temperature
			}
		}
	}

	return 0
}
