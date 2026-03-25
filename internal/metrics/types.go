package metrics

type Response struct {
	OK          bool            `json:"ok"`
	CollectedAt string          `json:"collectedAt,omitempty"`
	Overview    OverviewMetrics `json:"overview"`
	Network     NetworkMetrics  `json:"network"`
	Services    ServicesMetrics `json:"services"`
	Error       string          `json:"error,omitempty"`
}

type OverviewMetrics struct {
	CPUUsagePercent       float64 `json:"cpuUsagePercent"`
	CPUTemperatureCelsius float64 `json:"cpuTemperatureCelsius"`
	SSDTemperatureCelsius float64 `json:"ssdTemperatureCelsius"`
	RAMUsedBytes          uint64  `json:"ramUsedBytes"`
	RAMTotalBytes         uint64  `json:"ramTotalBytes"`
	RAMUsagePercent       float64 `json:"ramUsagePercent"`
	DiskUsedBytes         uint64  `json:"diskUsedBytes"`
	DiskTotalBytes        uint64  `json:"diskTotalBytes"`
	DiskUsagePercent      float64 `json:"diskUsagePercent"`
	UptimeSeconds         uint64  `json:"uptimeSeconds"`
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
