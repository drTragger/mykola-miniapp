package ups

import "time"

type BatteryResponse struct {
	OK             bool   `json:"ok"`
	CollectedAt    string `json:"collectedAt,omitempty"`
	BatteryPercent int    `json:"batteryPercent,omitempty"`
	Error          string `json:"error,omitempty"`
}

func GetBatterySnapshot() (BatteryResponse, error) {
	resp, err := GetSnapshot()
	if err != nil {
		return BatteryResponse{}, err
	}

	return BatteryResponse{
		OK:             true,
		CollectedAt:    resp.CollectedAt,
		BatteryPercent: resp.Data.BatteryPercent,
	}, nil
}

func GetBatterySnapshotFresh() (BatteryResponse, error) {
	resp, err := Collect()
	if err != nil {
		return BatteryResponse{}, err
	}

	return BatteryResponse{
		OK:             true,
		CollectedAt:    time.Now().Format(time.RFC3339),
		BatteryPercent: resp.Data.BatteryPercent,
	}, nil
}
