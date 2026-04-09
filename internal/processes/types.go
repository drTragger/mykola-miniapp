package processes

type ProcessInfo struct {
	PID           int32   `json:"pid"`
	PPID          int32   `json:"ppid"`
	Name          string  `json:"name"`
	User          string  `json:"user"`
	Status        string  `json:"status"`
	CPUPercent    float64 `json:"cpuPercent"`
	MemoryPercent float32 `json:"memoryPercent"`
	RSSBytes      uint64  `json:"rssBytes"`
	VMSBytes      uint64  `json:"vmsBytes"`
	Threads       int32   `json:"threads"`
	Cmdline       string  `json:"cmdline"`
	Executable    string  `json:"executable"`
	CreateTime    int64   `json:"createTime"`
	UptimeSec     int64   `json:"uptimeSec"`
	IsCritical    bool    `json:"isCritical"`
	ServiceName   string  `json:"serviceName,omitempty"`
}

type ListResponse struct {
	OK    bool          `json:"ok"`
	Items []ProcessInfo `json:"items"`
	Total int           `json:"total"`
	Error string        `json:"error,omitempty"`
}

type ActionRequest struct {
	Action string `json:"action"`
}

type ActionResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}
