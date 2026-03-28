package ups

type Response struct {
	OK            bool     `json:"ok"`
	CollectedAt   string   `json:"collectedAt,omitempty"`
	Data          Snapshot `json:"data"`
	Error         string   `json:"error,omitempty"`
	Stale         bool     `json:"stale"`
	LastSuccessAt string   `json:"lastSuccessAt,omitempty"`
}

type Snapshot struct {
	ModeText        string `json:"modeText"`
	PowerSourceText string `json:"powerSourceText"`
	ChargeText      string `json:"chargeText"`

	VBUSVoltageV float64 `json:"vbusVoltageV"`
	VBUSCurrentA float64 `json:"vbusCurrentA"`
	VBUSPowerW   float64 `json:"vbusPowerW"`

	BatteryVoltageV float64 `json:"batteryVoltageV"`
	BatteryCurrentA float64 `json:"batteryCurrentA"`
	BatteryPercent  int     `json:"batteryPercent"`
	RemainingMAh    int     `json:"remainingMAh"`
	FullCapacityMAh int     `json:"fullCapacityMAh"`

	Cell1MV int `json:"cell1Mv"`
	Cell2MV int `json:"cell2Mv"`
	Cell3MV int `json:"cell3Mv"`
	Cell4MV int `json:"cell4Mv"`

	CellDeltaMV   int    `json:"cellDeltaMv"`
	CellDeltaText string `json:"cellDeltaText"`

	TimeToChargeMin     int    `json:"timeToChargeMin"`
	TimeToChargeText    string `json:"timeToChargeText"`
	TimeToDischargeMin  int    `json:"timeToDischargeMin"`
	TimeToDischargeText string `json:"timeToDischargeText"`
	EtaText             string `json:"etaText"`

	CommText     string `json:"commText"`
	FirmwareText string `json:"firmwareText"`
}

type BatteryResponse struct {
	OK             bool   `json:"ok"`
	CollectedAt    string `json:"collectedAt,omitempty"`
	BatteryPercent int    `json:"batteryPercent,omitempty"`
	Error          string `json:"error,omitempty"`
	Stale          bool   `json:"stale"`
	LastSuccessAt  string `json:"lastSuccessAt,omitempty"`
}

type HistoryPoint struct {
	Timestamp      int64  `json:"timestamp"`
	BatteryPercent int    `json:"batteryPercent"`
	CellDeltaMV    int    `json:"cellDeltaMv"`
	Cell1MV        int    `json:"cell1Mv"`
	Cell2MV        int    `json:"cell2Mv"`
	Cell3MV        int    `json:"cell3Mv"`
	Cell4MV        int    `json:"cell4Mv"`
	Time           string `json:"time"`
}

type HistoryResponse struct {
	OK     bool           `json:"ok"`
	Points []HistoryPoint `json:"points"`
	Error  string         `json:"error,omitempty"`
}
