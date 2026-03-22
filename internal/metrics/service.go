package metrics

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	gnet "github.com/shirou/gopsutil/v3/net"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	netSampleMu   sync.Mutex
	lastSampleAt  time.Time
	lastRxBytes   uint64
	lastTxBytes   uint64
	httpClient    = &http.Client{Timeout: 2 * time.Second}
	serviceDialTO = 500 * time.Millisecond
)

func Collect() (Response, error) {
	var resp Response
	resp.OK = true
	resp.CollectedAt = time.Now().Format(time.RFC3339)

	var wg sync.WaitGroup

	wg.Add(4)

	go func() {
		defer wg.Done()
		fillOverviewAndSystem(&resp)
	}()

	go func() {
		defer wg.Done()
		fillNetwork(&resp)
	}()

	go func() {
		defer wg.Done()
		fillServices(&resp)
	}()

	go func() {
		defer wg.Done()
		fillVPN(&resp)
	}()

	wg.Wait()

	return resp, nil
}

func fillOverviewAndSystem(resp *Response) {
	vm, _ := mem.VirtualMemory()
	diskUsage, _ := disk.Usage("/")
	hostInfo, _ := host.Info()
	loadAvg, _ := load.Avg()

	cpuUsagePercent := 0.0
	if cpuPercents, err := cpu.Percent(0, false); err == nil && len(cpuPercents) > 0 {
		cpuUsagePercent = cpuPercents[0]
	}

	cpuModel := ""
	cpuFrequencyMHz := 0.0
	if cpuInfos, err := cpu.Info(); err == nil && len(cpuInfos) > 0 {
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
}

func fillNetwork(resp *Response) {
	localIPv4 := detectLocalIPv4()
	publicIP := detectPublicIP()
	pingMs := measureTCPPing("1.1.1.1:443")
	rxTotal, txTotal, rxSpeed, txSpeed := sampleNetworkTotals()

	resp.Network = NetworkMetrics{
		LocalIPv4:    localIPv4,
		PublicIP:     publicIP,
		PingMs:       pingMs,
		RxBytesTotal: rxTotal,
		TxBytesTotal: txTotal,
		RxSpeedBps:   rxSpeed,
		TxSpeedBps:   txSpeed,
		RxTotalHuman: humanBytes(rxTotal),
		TxTotalHuman: humanBytes(txTotal),
		RxSpeedHuman: humanBytes(uint64(rxSpeed)) + "/s",
		TxSpeedHuman: humanBytes(uint64(txSpeed)) + "/s",
	}
}

func fillServices(resp *Response) {
	resp.Services = ServicesMetrics{
		Jellyfin:    isTCPPortOpen("127.0.0.1:8096"),
		QBittorrent: isTCPPortOpen("127.0.0.1:8080"),
		Sonarr:      isTCPPortOpen("127.0.0.1:8989"),
		Radarr:      isTCPPortOpen("127.0.0.1:7878"),
		Prowlarr:    isTCPPortOpen("127.0.0.1:9696"),
		Fail2ban:    isServiceActive("fail2ban"),
	}
}

func fillVPN(resp *Response) {
	ok, ago := detectWireGuardHandshake()
	resp.VPN = VPNMetrics{
		OK:               ok,
		LastHandshakeAgo: ago,
	}
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

func detectPublicIP() string {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.ipify.org", nil)
	if err != nil {
		return "—"
	}

	resp, err := httpClient.Do(req)
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

func measureTCPPing(address string) float64 {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, 1500*time.Millisecond)
	if err != nil {
		return 0
	}
	_ = conn.Close()

	return float64(time.Since(start).Microseconds()) / 1000.0
}

func sampleNetworkTotals() (uint64, uint64, float64, float64) {
	stats, err := gnet.IOCounters(false)
	if err != nil || len(stats) == 0 {
		return 0, 0, 0, 0
	}

	rx := stats[0].BytesRecv
	tx := stats[0].BytesSent
	now := time.Now()

	netSampleMu.Lock()
	defer netSampleMu.Unlock()

	if lastSampleAt.IsZero() {
		lastSampleAt = now
		lastRxBytes = rx
		lastTxBytes = tx
		return rx, tx, 0, 0
	}

	elapsed := now.Sub(lastSampleAt).Seconds()
	if elapsed <= 0 {
		return rx, tx, 0, 0
	}

	rxSpeed := float64(rx-lastRxBytes) / elapsed
	txSpeed := float64(tx-lastTxBytes) / elapsed

	lastSampleAt = now
	lastRxBytes = rx
	lastTxBytes = tx

	return rx, tx, rxSpeed, txSpeed
}

func isTCPPortOpen(address string) bool {
	conn, err := net.DialTimeout("tcp", address, serviceDialTO)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func isServiceActive(name string) bool {
	out, err := exec.Command("systemctl", "is-active", name).Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(out)) == "active"
}

func detectWireGuardHandshake() (bool, string) {
	out, err := exec.Command("wg", "show", "all", "latest-handshakes").Output()
	if err != nil {
		return false, ""
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var latest int64

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		ts, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil || ts <= 0 {
			continue
		}

		if ts > latest {
			latest = ts
		}
	}

	if latest == 0 {
		return false, ""
	}

	diff := time.Since(time.Unix(latest, 0))
	if diff < 0 {
		diff = 0
	}

	return diff <= 3*time.Minute, humanAgo(diff)
}

func humanAgo(d time.Duration) string {
	if d < time.Minute {
		sec := int(d.Seconds())
		if sec < 1 {
			sec = 1
		}
		return fmt.Sprintf("%dс тому", sec)
	}

	if d < time.Hour {
		min := int(d.Minutes())
		return fmt.Sprintf("%dхв тому", min)
	}

	hours := int(d.Hours())
	return fmt.Sprintf("%dг тому", hours)
}

func humanBytes(v uint64) string {
	const unit = 1024
	if v < unit {
		return fmt.Sprintf("%d B", v)
	}

	div, exp := uint64(unit), 0
	for n := v / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	value := float64(v) / float64(div)
	suffixes := []string{"KiB", "MiB", "GiB", "TiB"}
	return fmt.Sprintf("%.2f %s", value, suffixes[exp])
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
