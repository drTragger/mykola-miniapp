package metrics

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	gnet "github.com/shirou/gopsutil/v3/net"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
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
	wg.Add(3)

	go func() {
		defer wg.Done()
		fillOverview(&resp)
	}()

	go func() {
		defer wg.Done()
		fillNetwork(&resp)
	}()

	go func() {
		defer wg.Done()
		fillServices(&resp)
	}()

	wg.Wait()

	return resp, nil
}

func fillOverview(resp *Response) {
	vm, _ := mem.VirtualMemory()
	hostInfo, _ := host.Info()

	cpuUsagePercent := 0.0
	if cpuPercents, err := cpu.Percent(0, false); err == nil && len(cpuPercents) > 0 {
		cpuUsagePercent = cpuPercents[0]
	}

	disks := collectDiskMetrics()

	var totalUsed uint64
	var totalSize uint64

	for _, diskItem := range disks {
		totalUsed += diskItem.UsedBytes
		totalSize += diskItem.TotalBytes
	}

	diskUsagePercent := 0.0
	if totalSize > 0 {
		diskUsagePercent = (float64(totalUsed) / float64(totalSize)) * 100
	}

	resp.Disks = disks
	resp.Overview = OverviewMetrics{
		CPUUsagePercent:       cpuUsagePercent,
		CPUTemperatureCelsius: readCPUTemperature(),
		SSDTemperatureCelsius: readSSDTemperature(),
		RAMUsedBytes:          vm.Used,
		RAMTotalBytes:         vm.Total,
		RAMUsagePercent:       vm.UsedPercent,
		DiskUsedBytes:         totalUsed,
		DiskTotalBytes:        totalSize,
		DiskUsagePercent:      diskUsagePercent,
		UptimeSeconds:         hostInfo.Uptime,
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

func collectDiskMetrics() []DiskMetrics {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil
	}

	items := make([]DiskMetrics, 0, len(partitions))
	seen := make(map[string]struct{})

	for _, p := range partitions {
		if !strings.HasPrefix(p.Device, "/dev/") {
			continue
		}

		if p.Mountpoint == "" {
			continue
		}

		key := p.Device + "|" + p.Mountpoint
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}

		parentDevice := detectParentBlockDevice(p.Device)
		if parentDevice == "" {
			parentDevice = p.Device
		}

		name := detectDiskName(parentDevice, p.Mountpoint)

		items = append(items, DiskMetrics{
			Name:               name,
			Device:             parentDevice,
			Mountpoint:         p.Mountpoint,
			Fstype:             p.Fstype,
			UsedBytes:          usage.Used,
			TotalBytes:         usage.Total,
			FreeBytes:          usage.Free,
			UsagePercent:       usage.UsedPercent,
			TemperatureCelsius: readDiskTemperature(parentDevice),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Mountpoint == "/" {
			return true
		}
		if items[j].Mountpoint == "/" {
			return false
		}
		return items[i].Mountpoint < items[j].Mountpoint
	})

	return items
}

func detectParentBlockDevice(source string) string {
	if source == "" || !strings.HasPrefix(source, "/dev/") {
		return ""
	}

	parent, err := runCommand(2, "lsblk", "-no", "PKNAME", source)
	if err == nil && strings.TrimSpace(parent) != "" {
		return "/dev/" + strings.TrimSpace(parent)
	}

	return source
}

func detectDiskName(device string, mountpoint string) string {
	if mountpoint == "/" {
		return "System SSD"
	}

	if mountpoint == "/data" {
		return "Data SSD"
	}

	label, err := runCommand(2, "lsblk", "-no", "LABEL", device)
	if err == nil && strings.TrimSpace(label) != "" {
		return strings.TrimSpace(label)
	}

	return strings.TrimPrefix(device, "/dev/")
}

func readDiskTemperature(device string) float64 {
	if device == "" {
		return 0
	}

	out, err := runSudoCommand(3, "smartctl", "-a", "-d", "sat", device)
	if err != nil {
		out, err = runSudoCommand(3, "smartctl", "-a", device)
		if err != nil {
			return 0
		}
	}

	re := regexp.MustCompile(`(?i)(Temperature_Celsius|Airflow_Temperature_Cel).*?(\d+)$`)

	for _, line := range strings.Split(out, "\n") {
		matches := re.FindStringSubmatch(strings.TrimSpace(line))
		if len(matches) == 3 {
			value, err := strconv.ParseFloat(matches[2], 64)
			if err == nil {
				return value
			}
		}
	}

	return 0
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

func readSSDTemperature() float64 {
	disks := collectDiskMetrics()
	if len(disks) == 0 {
		return 0
	}

	for _, diskItem := range disks {
		if diskItem.Mountpoint == "/" && diskItem.TemperatureCelsius > 0 {
			return diskItem.TemperatureCelsius
		}
	}

	for _, diskItem := range disks {
		if diskItem.TemperatureCelsius > 0 {
			return diskItem.TemperatureCelsius
		}
	}

	return 0
}

func detectRootDiskDevice() string {
	source, err := runCommand(2, "findmnt", "-n", "-o", "SOURCE", "/")
	if err != nil {
		return ""
	}

	source = strings.TrimSpace(source)
	if source == "" || !strings.HasPrefix(source, "/dev/") {
		return ""
	}

	parent, err := runCommand(2, "lsblk", "-no", "PKNAME", source)
	if err == nil && strings.TrimSpace(parent) != "" {
		return "/dev/" + strings.TrimSpace(parent)
	}

	return source
}

func runCommand(timeoutSec int, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("command timeout")
	}
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}

	return strings.TrimSpace(string(out)), nil
}

func runSudoCommand(timeoutSec int, cmd string, args ...string) (string, error) {
	allArgs := append([]string{"-n", cmd}, args...)
	return runCommand(timeoutSec, "sudo", allArgs...)
}
