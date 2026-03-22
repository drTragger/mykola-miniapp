package metrics

type Response struct {
	OK          bool            `json:"ok"`
	CollectedAt string          `json:"collectedAt,omitempty"`
	Overview    OverviewMetrics `json:"overview"`
	System      SystemMetrics   `json:"system"`
	Network     NetworkMetrics  `json:"network"`
	Services    ServicesMetrics `json:"services"`
	VPN         VPNMetrics      `json:"vpn"`
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

type ServicesMetrics struct {
	Jellyfin    bool `json:"jellyfin"`
	QBittorrent bool `json:"qBittorrent"`
	Sonarr      bool `json:"sonarr"`
	Radarr      bool `json:"radarr"`
	Prowlarr    bool `json:"prowlarr"`
	Fail2ban    bool `json:"fail2ban"`
}

type VPNMetrics struct {
	OK               bool   `json:"ok"`
	Status           bool   `json:"status"`
	RouteOK          bool   `json:"routeOk"`
	LastHandshakeAgo string `json:"lastHandshakeAgo"`
}
