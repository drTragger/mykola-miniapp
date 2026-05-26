package system

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"

	"github.com/drTragger/mykola-miniapp/internal/sysutil"
)

var netSampler = sysutil.NewNetworkSampler()

func Collect() (Response, error) {
	resp := Response{
		OK:          true,
		CollectedAt: time.Now().Format(time.RFC3339),
	}

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
	var (
		hostInfo        *host.InfoStat
		loadAvg         *load.AvgStat
		cpuInfos        []cpu.InfoStat
		logicalCPUCount int
		hostname        string
	)

	var wg sync.WaitGroup
	wg.Add(5)

	go func() {
		defer wg.Done()
		hostInfo, _ = host.Info()
	}()

	go func() {
		defer wg.Done()
		loadAvg, _ = load.Avg()
	}()

	go func() {
		defer wg.Done()
		cpuInfos, _ = cpu.Info()
	}()

	go func() {
		defer wg.Done()
		logicalCPUCount, _ = cpu.Counts(true)
	}()

	go func() {
		defer wg.Done()
		hostname, _ = os.Hostname()
	}()

	wg.Wait()

	if hostInfo == nil {
		hostInfo = &host.InfoStat{}
	}
	if loadAvg == nil {
		loadAvg = &load.AvgStat{}
	}

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

	rxSpeedHuman := "—"
	txSpeedHuman := "—"
	if rxSpeed > 0 {
		rxSpeedHuman = sysutil.HumanBytes(uint64(rxSpeed)) + "/s"
	}
	if txSpeed > 0 {
		txSpeedHuman = sysutil.HumanBytes(uint64(txSpeed)) + "/s"
	}

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
		RxSpeedHuman: rxSpeedHuman,
		TxSpeedHuman: txSpeedHuman,
	}
}

func fillVPN(resp *Response) {
	var (
		serviceOK         bool
		wgIP              string
		endpoint          string
		handshakeAgo      string
		rx, tx            string
		routeTable        string
		qbitServiceStatus bool
		qbitServiceUser   string
		qbitBinding       string
		qbitWebUI         string
		ruleOK            bool
		routeOK           bool
	)

	var wg sync.WaitGroup
	wg.Add(8)

	go func() {
		defer wg.Done()
		serviceOK = sysutil.IsServiceActive("wg-quick@wg0")
	}()

	go func() {
		defer wg.Done()
		wgIP = getWGInterfaceIP()
	}()

	go func() {
		defer wg.Done()
		endpoint, handshakeAgo, rx, tx = getWGDumpData()
	}()

	go func() {
		defer wg.Done()
		routeTable = getVPNRouteTable()
	}()

	go func() {
		defer wg.Done()
		qbitServiceStatus = getQBittorrentServiceStatus()
	}()

	go func() {
		defer wg.Done()
		qbitServiceUser = getQBittorrentServiceUser()
	}()

	go func() {
		defer wg.Done()
		qbitBinding, qbitWebUI = readQBittorrentBindingAndWebUI()
	}()

	go func() {
		defer wg.Done()
		ruleOK = hasVPNRuleForQBittorrent()
		routeOK = hasVPNRoute()
	}()

	wg.Wait()

	bindingOK := qbitBinding == "wg0"
	handshakeOK := handshakeAgo != "ніколи" && handshakeAgo != "н/д"
	ok := serviceOK && handshakeOK && ruleOK && qbitServiceStatus && bindingOK

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

func hasVPNRoute() bool {
	checks := [][]string{
		{"ip", "route", "show", "table", "vpn"},
		{"ip", "route", "show", "table", "51820"},
		{"ip", "route"},
	}

	for _, cmd := range checks {
		out, err := sysutil.RunCommand(2, cmd[0], cmd[1:]...)
		if err != nil {
			continue
		}

		if strings.Contains(out, "default dev wg0") || strings.Contains(out, "10.2.0.1 dev wg0") {
			return true
		}
	}

	return false
}

func fillUsers(resp *Response) {
	users, err := getLoggedInUsers()
	if err != nil {
		resp.Users = []UserSession{}
		return
	}

	resp.Users = users
}

func getWGInterfaceIP() string {
	out, err := sysutil.RunCommand(2, "ip", "-4", "-o", "addr", "show", "dev", "wg0")
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
	out, err := sysutil.RunSudoCommand(2, "wg", "show", "wg0", "dump")
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
		handshakeAgo = sysutil.HumanizeDurationShort(time.Since(time.Unix(handshakeUnix, 0))) + " тому"
	}

	rxBytes, err := strconv.ParseUint(fields[5], 10, 64)
	if err != nil {
		rx = "н/д"
	} else {
		rx = sysutil.HumanBytes(rxBytes)
	}

	txBytes, err := strconv.ParseUint(fields[6], 10, 64)
	if err != nil {
		tx = "н/д"
	} else {
		tx = sysutil.HumanBytes(txBytes)
	}

	return endpoint, handshakeAgo, rx, tx
}

func hasVPNRuleForQBittorrent() bool {
	_, uid := getQBittorrentUserInfo()
	if uid == "" {
		return false
	}

	out, err := sysutil.RunCommand(2, "ip", "rule")
	if err != nil {
		return false
	}

	expected := fmt.Sprintf("uidrange %s-%s lookup vpn", uid, uid)
	return strings.Contains(out, expected)
}

func getVPNRouteTable() string {
	checks := []struct {
		name string
		args []string
	}{
		{name: "vpn", args: []string{"route", "show", "table", "vpn"}},
		{name: "51820", args: []string{"route", "show", "table", "51820"}},
	}

	for _, check := range checks {
		out, err := sysutil.RunCommand(2, "ip", check.args...)
		if err == nil && strings.TrimSpace(out) != "" {
			return strings.TrimSpace(out)
		}
	}

	out, err := sysutil.RunCommand(2, "ip", "route")
	if err == nil && strings.TrimSpace(out) != "" {
		var lines []string
		for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
			if strings.Contains(line, "wg0") {
				lines = append(lines, line)
			}
		}
		if len(lines) > 0 {
			return strings.Join(lines, "\n")
		}
	}

	return "❌ відсутній"
}

func getQBittorrentUserInfo() (username string, uid string) {
	u, err := user.Lookup("qbittorrent")
	if err != nil {
		return "qbittorrent", ""
	}

	return u.Username, u.Uid
}

var qbittorrentServiceNames = []string{
	"qbittorrent",
	"qbittorrent.service",
	"qbittorrent-nox",
	"qbittorrent-nox.service",
}

func getQBittorrentServiceUser() string {
	for _, service := range qbittorrentServiceNames {
		out, err := sysutil.RunCommand(2, "systemctl", "show", service, "--property=User", "--value")
		if err == nil && strings.TrimSpace(out) != "" {
			return strings.TrimSpace(out)
		}
	}

	return "н/д"
}

func getQBittorrentServiceStatus() bool {
	for _, service := range qbittorrentServiceNames {
		if sysutil.IsServiceActive(service) {
			return true
		}
	}
	return false
}

func readQBittorrentBindingAndWebUI() (binding string, webui string) {
	data := readQBittorrentConfig()
	if data == "" {
		return "н/д", "н/д"
	}

	binding = "н/д"
	address := "0.0.0.0"
	port := "8080"

	bindingKeys := []string{
		"Connection\\Interface=",
		"Connection\\InterfaceName=",
		"Session\\Interface=",
		"Session\\InterfaceName=",
	}

	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)

		if binding == "н/д" {
			for _, key := range bindingKeys {
				if strings.HasPrefix(line, key) {
					binding = strings.TrimSpace(strings.TrimPrefix(line, key))
					break
				}
			}
		}

		if strings.HasPrefix(line, "WebUI\\Address=") {
			address = strings.TrimSpace(strings.TrimPrefix(line, "WebUI\\Address="))
		}

		if strings.HasPrefix(line, "WebUI\\Port=") {
			port = strings.TrimSpace(strings.TrimPrefix(line, "WebUI\\Port="))
		}
	}

	return binding, fmt.Sprintf("%s:%s", address, port)
}

func readQBittorrentConfig() string {
	paths := []string{
		"/home/qbittorrent/.config/qBittorrent/qBittorrent.conf",
		"/home/qbittorrent/.local/share/qBittorrent/qBittorrent.conf",
		"/home/qbittorrent/.config/qBittorrent/config/qBittorrent.conf",
		"/home/qbittorrent/.config/qBittorrent/qBittorrent-data.conf",
	}

	for _, path := range paths {
		out, err := sysutil.RunSudoCommand(2, "cat", path)
		if err == nil && strings.TrimSpace(out) != "" {
			return out
		}
	}

	return ""
}

func getLoggedInUsers() ([]UserSession, error) {
	out, err := sysutil.RunCommand(3, "who")
	if err != nil || strings.TrimSpace(out) == "" {
		return []UserSession{}, nil
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	result := make([]UserSession, 0, len(lines))

	idleMap := getIdleMap()

	for _, line := range lines {
		session, ok := parseWhoLine(line)
		if !ok {
			continue
		}

		if idle, ok := idleMap[session.TTY]; ok {
			session.Idle = idle
		}

		session.Location = detectLocationLabel(session.RemoteIP, session.ConnectionType)
		result = append(result, session)
	}

	return result, nil
}

func parseWhoLine(line string) (UserSession, bool) {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return UserSession{}, false
	}

	username := fields[0]
	tty := fields[1]
	datePart := fields[2]
	timePart := fields[3]

	from := ""
	if len(fields) >= 5 {
		from = strings.Trim(fields[4], "()")
	}

	loginAt := fmt.Sprintf("%s %s", datePart, timePart)
	connectionType := detectConnectionType(tty, from)

	return UserSession{
		Username:       username,
		TTY:            tty,
		From:           from,
		RemoteIP:       normalizeRemoteIP(from),
		LoginAt:        loginAt,
		Idle:           "—",
		ConnectionType: connectionType,
		IsLocal:        isLocalSession(tty, from),
		Location:       "—",
	}, true
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

	if strings.HasPrefix(tty, "pts/") {
		if from != "" && from != ":0" && from != "localhost" {
			return false
		}
		return true
	}

	if from == "" || from == ":0" || from == "localhost" {
		return true
	}

	return false
}

func normalizeRemoteIP(from string) string {
	if ip := net.ParseIP(from); ip != nil {
		return ip.String()
	}

	return from
}

func getIdleMap() map[string]string {
	out, err := sysutil.RunCommand(3, "who", "-u")
	if err != nil || strings.TrimSpace(out) == "" {
		return map[string]string{}
	}

	result := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		tty := fields[1]
		idle := normalizeIdle(fields[4])

		if tty != "" {
			result[tty] = idle
		}
	}

	return result
}

func normalizeIdle(idle string) string {
	idle = strings.TrimSpace(idle)

	switch idle {
	case "", "—":
		return "—"
	case ".":
		return "активний"
	case "old":
		return "давно неактивний"
	default:
		return idle
	}
}

func detectLocationLabel(remoteIP, connectionType string) string {
	if connectionType == "local" {
		return "Ця машина"
	}

	if remoteIP == "" || remoteIP == ":0" || remoteIP == "localhost" {
		return "Локально"
	}

	ip := net.ParseIP(remoteIP)
	if ip == nil {
		return "Віддалений хост"
	}

	if ip.IsLoopback() {
		return "Локально"
	}

	if isPrivateOrLocalIP(ip) {
		return "Локальна / VPN мережа"
	}

	return "Публічна мережа"
}

func isPrivateOrLocalIP(ip net.IP) bool {
	return ip.IsPrivate() ||
		ip.IsLoopback() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast()
}
