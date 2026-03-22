package system

type Response struct {
	OK          bool           `json:"ok"`
	CollectedAt string         `json:"collectedAt,omitempty"`
	System      SystemMetrics  `json:"system"`
	Network     NetworkMetrics `json:"network"`
	VPN         VPNMetrics     `json:"vpn"`
	Error       string         `json:"error,omitempty"`
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
	LocalIPv4    string  `json:"localIpv4"`
	PublicIP     string  `json:"publicIp"`
	PingMs       float64 `json:"pingMs"`
	RxBytesTotal uint64  `json:"rxBytesTotal"`
	TxBytesTotal uint64  `json:"txBytesTotal"`
	RxSpeedBps   float64 `json:"rxSpeedBps"`
	TxSpeedBps   float64 `json:"txSpeedBps"`
	RxTotalHuman string  `json:"rxTotalHuman"`
	TxTotalHuman string  `json:"txTotalHuman"`
	RxSpeedHuman string  `json:"rxSpeedHuman"`
	TxSpeedHuman string  `json:"txSpeedHuman"`
}

type VPNMetrics struct {
	OK               bool           `json:"ok"`
	ServiceOK        bool           `json:"serviceOk"`
	WgIP             string         `json:"wgIp"`
	Endpoint         string         `json:"endpoint"`
	LastHandshakeAgo string         `json:"lastHandshakeAgo"`
	RX               string         `json:"rx"`
	TX               string         `json:"tx"`
	RuleOK           bool           `json:"ruleOk"`
	RouteOK          bool           `json:"routeOk"`
	RouteTable       string         `json:"routeTable"`
	QBit             QBittorrentVPN `json:"qbit"`
}

type QBittorrentVPN struct {
	ServiceOK bool   `json:"serviceOk"`
	User      string `json:"user"`
	Binding   string `json:"binding"`
	WebUI     string `json:"webui"`
}
