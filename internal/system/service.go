package system

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	gnet "github.com/shirou/gopsutil/v3/net"
)

var (
	netSampleMu  sync.Mutex
	lastSampleAt time.Time
	lastRxBytes  uint64
	lastTxBytes  uint64
	httpClient   = &http.Client{Timeout: 2 * time.Second}
)

func Collect() (Response, error) {
	var resp Response
	resp.OK = true
	resp.CollectedAt = time.Now().Format(time.RFC3339)

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		fillSystem(&resp)
	}()

	go func() {
		defer wg.Done()
		fillNetwork(&resp)
	}()

	go func() {
		defer wg.Done()
		fillVPN(&resp)
	}()

	go func() {
		defer wg.Done()
		fillUsers(&resp)
	}()

	wg.Wait()

	return resp, nil
}

func fillSystem(resp *Response) {
	hostInfo, _ := host.Info()
	loadAvg, _ := load.Avg()
	cpuInfos, _ := cpu.Info()
	logicalCPUCount, _ := cpu.Counts(true)
	hostname, _ := os.Hostname()

	cpuModel := ""
	cpuFrequencyMHz := 0.0
	if len(cpuInfos) > 0 {
		cpuModel = cpuInfos[0].ModelName
		cpuFrequencyMHz = cpuInfos[0].Mhz
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

func fillVPN(resp *Response) {
	serviceOK := isServiceActive("wg-quick@wg0")
	wgIP := getWGInterfaceIP()
	endpoint, handshakeAgo, rx, tx := getWGDumpData()
	routeTable := getVPNRouteTable()

	qbitServiceStatus := getQBittorrentServiceStatus()
	qbitServiceUser := getQBittorrentServiceUser()
	qbitBinding := getQBittorrentInterfaceBinding()
	qbitWebUI := getQBittorrentWebUIAddress()

	ruleOK := hasVPNRuleForQBittorrent()
	routeOK := strings.Contains(routeTable, "default dev wg0")
	ok := serviceOK && routeOK && handshakeAgo != "ніколи" && handshakeAgo != "н/д"

	resp.VPN = VPNMetrics{
		OK:               ok,
		ServiceOK:        serviceOK,
		WgIP:             wgIP,
		Endpoint:         endpoint,
		LastHandshakeAgo: handshakeAgo,
		RX:               rx,
		TX:               tx,
		RuleOK:           ruleOK,
		RouteOK:          routeOK,
		RouteTable:       routeTable,
		QBit: QBittorrentVPN{
			ServiceOK: qbitServiceStatus,
			User:      qbitServiceUser,
			Binding:   qbitBinding,
			WebUI:     qbitWebUI,
		},
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

func humanizeDurationShort(d time.Duration) string {
	if d < 0 {
		return "н/д"
	}

	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%dс", seconds)
	}

	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dхв", minutes)
	}

	hours := minutes / 60
	minutes = minutes % 60
	if hours < 24 {
		return fmt.Sprintf("%dг %dхв", hours, minutes)
	}

	days := hours / 24
	hours = hours % 24
	return fmt.Sprintf("%dд %dг", days, hours)
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

func getWGInterfaceIP() string {
	out, err := runCommand(2, "ip", "-4", "-o", "addr", "show", "dev", "wg0")
	if err != nil || out == "" {
		return "н/д"
	}

	fields := strings.Fields(out)
	if len(fields) < 4 {
		return "н/д"
	}

	return fields[3]
}

func getWGDumpData() (endpoint, handshakeAgo, rx, tx string) {
	out, err := runSudoCommand(2, "wg", "show", "wg0", "dump")
	if err != nil || out == "" {
		return "н/д", "н/д", "н/д", "н/д"
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		return "н/д", "н/д", "н/д", "н/д"
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 8 {
		return "н/д", "н/д", "н/д", "н/д"
	}

	endpoint = fields[2]

	handshakeUnix, err := strconv.ParseInt(fields[4], 10, 64)
	if err != nil || handshakeUnix == 0 {
		handshakeAgo = "ніколи"
	} else {
		handshakeAgo = humanizeDurationShort(time.Since(time.Unix(handshakeUnix, 0))) + " тому"
	}

	rxBytes, err := strconv.ParseUint(fields[5], 10, 64)
	if err != nil {
		rx = "н/д"
	} else {
		rx = humanBytes(rxBytes)
	}

	txBytes, err := strconv.ParseUint(fields[6], 10, 64)
	if err != nil {
		tx = "н/д"
	} else {
		tx = humanBytes(txBytes)
	}

	return endpoint, handshakeAgo, rx, tx
}

func hasVPNRuleForQBittorrent() bool {
	_, uid := getQBittorrentUserInfo()
	if uid == "" {
		return false
	}

	out, err := runCommand(2, "ip", "rule")
	if err != nil {
		return false
	}

	expected := fmt.Sprintf("uidrange %s-%s lookup vpn", uid, uid)
	return strings.Contains(out, expected)
}

func getVPNRouteTable() string {
	out, err := runCommand(2, "ip", "route", "show", "table", "vpn")
	if err != nil || strings.TrimSpace(out) == "" {
		return "❌ відсутній"
	}

	return strings.TrimSpace(out)
}

func getQBittorrentUserInfo() (username string, uid string) {
	u, err := user.Lookup("qbittorrent")
	if err != nil {
		return "qbittorrent", ""
	}

	return u.Username, u.Uid
}

func getQBittorrentServiceUser() string {
	serviceNames := []string{
		"qbittorrent",
		"qbittorrent.service",
		"qbittorrent-nox",
		"qbittorrent-nox.service",
	}

	for _, service := range serviceNames {
		out, err := runCommand(2, "systemctl", "show", service, "--property=User", "--value")
		if err == nil && strings.TrimSpace(out) != "" {
			return strings.TrimSpace(out)
		}
	}

	return "н/д"
}

func getQBittorrentServiceStatus() bool {
	serviceNames := []string{
		"qbittorrent",
		"qbittorrent.service",
		"qbittorrent-nox",
		"qbittorrent-nox.service",
	}

	for _, service := range serviceNames {
		if isServiceActive(service) {
			return true
		}
	}

	return false
}

func getQBittorrentInterfaceBinding() string {
	data := readQBittorrentConfig()
	if data == "" {
		return "н/д"
	}

	lines := strings.Split(data, "\n")
	keys := []string{
		"Connection\\Interface=",
		"Connection\\InterfaceName=",
		"Session\\Interface=",
		"Session\\InterfaceName=",
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		for _, key := range keys {
			if strings.HasPrefix(line, key) {
				return strings.TrimSpace(strings.TrimPrefix(line, key))
			}
		}
	}

	return "н/д"
}

func getQBittorrentWebUIAddress() string {
	data := readQBittorrentConfig()
	if data == "" {
		return "н/д"
	}

	lines := strings.Split(data, "\n")
	address := "0.0.0.0"
	port := "8080"

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "WebUI\\Address=") {
			address = strings.TrimSpace(strings.TrimPrefix(line, "WebUI\\Address="))
		}

		if strings.HasPrefix(line, "WebUI\\Port=") {
			port = strings.TrimSpace(strings.TrimPrefix(line, "WebUI\\Port="))
		}
	}

	return fmt.Sprintf("%s:%s", address, port)
}

func readQBittorrentConfig() string {
	paths := []string{
		"/home/qbittorrent/.config/qBittorrent/qBittorrent.conf",
		"/home/qbittorrent/.local/share/qBittorrent/qBittorrent.conf",
		"/home/qbittorrent/.config/qBittorrent/config/qBittorrent.conf",
		"/home/qbittorrent/.config/qBittorrent/qBittorrent-data.conf",
	}

	for _, path := range paths {
		out, err := runSudoCommand(2, "cat", path)
		if err == nil && strings.TrimSpace(out) != "" {
			return out
		}
	}

	return ""
}

func isServiceActive(name string) bool {
	out, err := exec.Command("systemctl", "is-active", name).Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(out)) == "active"
}

func fillUsers(resp *Response) {
	users, err := getLoggedInUsers()
	if err != nil {
		resp.Users = []UserSession{}
		return
	}

	resp.Users = users
}

func getLoggedInUsers() ([]UserSession, error) {
	out, err := runCommand(3, "who", "--ips")
	if err != nil || strings.TrimSpace(out) == "" {
		return []UserSession{}, nil
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	result := make([]UserSession, 0, len(lines))

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		username := fields[0]
		tty := fields[1]
		datePart := fields[2]
		timePart := fields[3]
		from := strings.Trim(fields[4], "()")

		loginAt := fmt.Sprintf("%s %s", datePart, timePart)

		session := UserSession{
			Username:       username,
			TTY:            tty,
			From:           from,
			RemoteIP:       normalizeRemoteIP(from),
			LoginAt:        loginAt,
			Idle:           "—",
			ConnectionType: detectConnectionType(tty, from),
			IsLocal:        isLocalSession(tty, from),
			Location:       "—",
		}

		result = append(result, session)
	}

	idleMap := getIdleMap()

	for i := range result {
		if idle, ok := idleMap[result[i].TTY]; ok {
			result[i].Idle = idle
		}
	}

	return result, nil
}

func detectConnectionType(tty, from string) string {
	if strings.HasPrefix(tty, "tty") {
		return "local"
	}

	if strings.HasPrefix(tty, "pts/") {
		if from != "" && from != ":0" && from != "localhost" {
			return "ssh"
		}
		return "terminal"
	}

	return "unknown"
}

func isLocalSession(tty, from string) bool {
	if strings.HasPrefix(tty, "tty") {
		return true
	}

	if from == "" || from == ":0" || from == "localhost" {
		return true
	}

	ip := net.ParseIP(from)
	if ip == nil {
		return false
	}

	return ip.IsLoopback() || ip.IsPrivate()
}

func normalizeRemoteIP(from string) string {
	ip := net.ParseIP(from)
	if ip != nil {
		return ip.String()
	}

	return from
}

func getIdleMap() map[string]string {
	out, err := runCommand(3, "w", "-h")
	if err != nil || strings.TrimSpace(out) == "" {
		return map[string]string{}
	}

	result := map[string]string{}
	lines := strings.Split(strings.TrimSpace(out), "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		tty := fields[1]
		idle := fields[4]
		result[tty] = idle
	}

	return result
}
