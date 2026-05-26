package metrics

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/drTragger/mykola-miniapp/internal/sysutil"
)

const serviceDialTO = 500 * time.Millisecond

var netSampler = sysutil.NewNetworkSampler()

func Collect() (Response, error) {
	resp := Response{
		OK:          true,
		CollectedAt: time.Now().Format(time.RFC3339),
	}

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
	var (
		vm              *mem.VirtualMemoryStat
		hostInfo        *host.InfoStat
		cpuUsagePercent float64
		disks           []DiskMetrics
	)

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		vm, _ = mem.VirtualMemory()
	}()

	go func() {
		defer wg.Done()
		hostInfo, _ = host.Info()
	}()

	go func() {
		defer wg.Done()
		if cpuPercents, err := cpu.Percent(0, false); err == nil && len(cpuPercents) > 0 {
			cpuUsagePercent = cpuPercents[0]
		}
	}()

	go func() {
		defer wg.Done()
		disks = collectDiskMetrics()
	}()

	wg.Wait()

	var totalUsed, totalSize uint64
	for _, d := range disks {
		if d.Mountpoint != "/" && d.Mountpoint != "/data" {
			continue
		}
		totalUsed += d.UsedBytes
		totalSize += d.TotalBytes
	}

	diskUsagePercent := 0.0
	if totalSize > 0 {
		diskUsagePercent = (float64(totalUsed) / float64(totalSize)) * 100
	}

	if vm == nil {
		vm = &mem.VirtualMemoryStat{}
	}
	if hostInfo == nil {
		hostInfo = &host.InfoStat{}
	}

	resp.Disks = disks
	resp.Overview = OverviewMetrics{
		CPUUsagePercent:       cpuUsagePercent,
		CPUTemperatureCelsius: readCPUTemperature(),
		SSDTemperatureCelsius: readSSDTemperature(disks),
		CPUThrottled:          isCPUThrottled(),
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
	var (
		localIPv4 string
		publicIP  string
		pingMs    float64
	)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		localIPv4 = sysutil.DetectLocalIPv4()
	}()

	go func() {
		defer wg.Done()
		publicIP = sysutil.DetectPublicIP()
	}()

	go func() {
		defer wg.Done()
		pingMs = sysutil.MeasureTCPPing("1.1.1.1:443")
	}()

	rxTotal, txTotal, rxSpeed, txSpeed := netSampler.Sample()

	wg.Wait()

	resp.Network = NetworkMetrics{
		LocalIPv4:    localIPv4,
		PublicIP:     publicIP,
		PingMs:       pingMs,
		RxBytesTotal: rxTotal,
		TxBytesTotal: txTotal,
		RxSpeedBps:   rxSpeed,
		TxSpeedBps:   txSpeed,
		RxTotalHuman: sysutil.HumanBytes(rxTotal),
		TxTotalHuman: sysutil.HumanBytes(txTotal),
		RxSpeedHuman: sysutil.HumanBytes(uint64(rxSpeed)) + "/s",
		TxSpeedHuman: sysutil.HumanBytes(uint64(txSpeed)) + "/s",
	}
}

func fillServices(resp *Response) {
	targets := []struct {
		name string
		addr string
		dst  *bool
	}{
		{"Jellyfin", "127.0.0.1:8096", &resp.Services.Jellyfin},
		{"QBittorrent", "127.0.0.1:8080", &resp.Services.QBittorrent},
		{"Sonarr", "127.0.0.1:8989", &resp.Services.Sonarr},
		{"Radarr", "127.0.0.1:7878", &resp.Services.Radarr},
		{"Prowlarr", "127.0.0.1:9696", &resp.Services.Prowlarr},
	}

	var wg sync.WaitGroup
	wg.Add(len(targets) + 1)

	for _, t := range targets {
		t := t
		go func() {
			defer wg.Done()
			*t.dst = sysutil.IsTCPPortOpen(t.addr, serviceDialTO)
		}()
	}

	go func() {
		defer wg.Done()
		resp.Services.Fail2ban = sysutil.IsServiceActive("fail2ban")
	}()

	wg.Wait()
}

func collectDiskMetrics() []DiskMetrics {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil
	}

	allowedMountpoints := map[string]bool{
		"/":       true,
		"/data":   true,
		"/backup": true,
	}

	type entry struct {
		device     string
		mountpoint string
		fstype     string
	}

	candidates := make([]entry, 0, len(partitions))
	seen := make(map[string]struct{})

	for _, p := range partitions {
		if !strings.HasPrefix(p.Device, "/dev/") || p.Mountpoint == "" {
			continue
		}
		if !allowedMountpoints[p.Mountpoint] {
			continue
		}

		key := p.Device + "|" + p.Mountpoint
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		candidates = append(candidates, entry{p.Device, p.Mountpoint, p.Fstype})
	}

	items := make([]DiskMetrics, len(candidates))

	var wg sync.WaitGroup
	for i, c := range candidates {
		i, c := i, c
		wg.Add(1)
		go func() {
			defer wg.Done()

			usage, err := disk.Usage(c.mountpoint)
			if err != nil {
				return
			}

			parentDevice := detectParentBlockDevice(c.device)
			if parentDevice == "" {
				parentDevice = c.device
			}

			items[i] = DiskMetrics{
				Name:               detectDiskName(parentDevice, c.mountpoint),
				Device:             parentDevice,
				Mountpoint:         c.mountpoint,
				Fstype:             c.fstype,
				UsedBytes:          usage.Used,
				TotalBytes:         usage.Total,
				FreeBytes:          usage.Free,
				UsagePercent:       usage.UsedPercent,
				TemperatureCelsius: readDiskTemperature(parentDevice),
			}
		}()
	}
	wg.Wait()

	filtered := items[:0]
	for _, it := range items {
		if it.Device != "" {
			filtered = append(filtered, it)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		order := func(m string) int {
			switch m {
			case "/":
				return 0
			case "/data":
				return 1
			case "/backup":
				return 2
			default:
				return 99
			}
		}
		return order(filtered[i].Mountpoint) < order(filtered[j].Mountpoint)
	})

	return filtered
}

func detectParentBlockDevice(source string) string {
	if source == "" || !strings.HasPrefix(source, "/dev/") {
		return ""
	}

	parent, err := sysutil.RunCommand(2, "lsblk", "-no", "PKNAME", source)
	if err == nil && strings.TrimSpace(parent) != "" {
		return "/dev/" + strings.TrimSpace(parent)
	}

	return source
}

func detectDiskName(device string, mountpoint string) string {
	switch mountpoint {
	case "/":
		return "System SSD"
	case "/data":
		return "Media SSD"
	case "/backup":
		return "Backup SSD"
	}

	label, err := sysutil.RunCommand(2, "lsblk", "-no", "LABEL", device)
	if err == nil && strings.TrimSpace(label) != "" {
		return strings.TrimSpace(label)
	}

	return strings.TrimPrefix(device, "/dev/")
}

func readDiskTemperature(device string) float64 {
	if device == "" {
		return 0
	}

	if strings.HasPrefix(device, "/dev/nvme") {
		out, err := runSmartctlCommand(5, device)
		if err != nil && strings.TrimSpace(out) == "" {
			log.Printf("nvme smartctl error for %s: %v", device, err)
			return 0
		}

		return parseNvmeTemperature(out)
	}

	out, err := runSmartctlCommand(5, "-d", "sat", device)
	if err != nil && strings.TrimSpace(out) == "" {
		out, err = runSmartctlCommand(5, device)
		if err != nil && strings.TrimSpace(out) == "" {
			log.Printf("ata smartctl error for %s: %v", device, err)
			return 0
		}
	}

	return parseAtaTemperature(out)
}

func parseNvmeTemperature(out string) float64 {
	prefixes := []string{"Temperature:", "Temperature Sensor 1:", "Temperature Sensor 2:"}

	for _, prefix := range prefixes {
		if temp := findTempInLines(out, prefix); temp > 0 {
			return temp
		}
	}

	return 0
}

func findTempInLines(out, prefix string) float64 {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		for _, field := range strings.Fields(line) {
			if value, err := strconv.ParseFloat(field, 64); err == nil {
				return value
			}
		}
	}
	return 0
}

func parseAtaTemperature(out string) float64 {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)

		if len(fields) < 10 {
			continue
		}

		if fields[0] == "194" && fields[1] == "Temperature_Celsius" ||
			fields[0] == "190" && fields[1] == "Airflow_Temperature_Cel" {
			if value, err := strconv.ParseFloat(fields[len(fields)-1], 64); err == nil {
				return value
			}
		}
	}

	return 0
}

func readCPUTemperature() float64 {
	if raw, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp"); err == nil {
		if value, parseErr := strconv.ParseFloat(strings.TrimSpace(string(raw)), 64); parseErr == nil {
			return value / 1000.0
		}
	}

	return 0
}

func isCPUThrottled() bool {
	out, err := sysutil.RunCommand(2, "vcgencmd", "get_throttled")
	if err != nil {
		return false
	}

	parts := strings.Split(strings.TrimSpace(out), "=")
	if len(parts) != 2 {
		return false
	}

	valueStr := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(parts[1])), "0x")

	value, err := strconv.ParseUint(valueStr, 16, 64)
	if err != nil {
		return false
	}

	const mask uint64 = (1 << 0) | (1 << 1) | (1 << 2) | (1 << 3)
	return value&mask != 0
}

func readSSDTemperature(disks []DiskMetrics) float64 {
	if len(disks) == 0 {
		return 0
	}

	for _, d := range disks {
		if d.Mountpoint == "/" && d.TemperatureCelsius > 0 {
			return d.TemperatureCelsius
		}
	}

	for _, d := range disks {
		if d.TemperatureCelsius > 0 {
			return d.TemperatureCelsius
		}
	}

	return 0
}

func runSmartctlCommand(timeoutSec int, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	allArgs := append([]string{"-n", "smartctl", "-a"}, args...)
	cmd := exec.CommandContext(ctx, "sudo", allArgs...)
	out, err := cmd.CombinedOutput()

	text := strings.TrimSpace(string(out))

	if ctx.Err() == context.DeadlineExceeded {
		return text, fmt.Errorf("smartctl timeout")
	}

	if err != nil {
		if text != "" {
			return text, fmt.Errorf("%s", text)
		}
		return "", err
	}

	return text, nil
}
