package ups

func GetBatterySnapshot() (BatteryResponse, error) {
	resp, err := GetSnapshot()
	if err != nil {
		return BatteryResponse{}, err
	}

	return BatteryResponse{
		OK:             resp.OK,
		CollectedAt:    resp.CollectedAt,
		BatteryPercent: resp.Data.BatteryPercent,
		Error:          resp.Error,
		Stale:          resp.Stale,
		LastSuccessAt:  resp.LastSuccessAt,
	}, nil
}

func GetBatterySnapshotFresh() (BatteryResponse, error) {
	resp, err := Collect()
	if err != nil {
		return BatteryResponse{}, err
	}

	return BatteryResponse{
		OK:             true,
		CollectedAt:    resp.CollectedAt,
		BatteryPercent: resp.Data.BatteryPercent,
	}, nil
}
