package ups

type Response struct {
	OK          bool     `json:"ok"`
	CollectedAt string   `json:"collectedAt,omitempty"`
	Data        Snapshot `json:"data"`
	Error       string   `json:"error,omitempty"`
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

	CommText     string `json:"commText"`
	FirmwareText string `json:"firmwareText"`
}
