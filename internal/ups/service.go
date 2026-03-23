package ups

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	upsI2CBus  = 1
	upsI2CAddr = "0x2d"

	regChargeState = "0x02"
	regCommState   = "0x03"

	regVBUSVoltageLo = "0x10"
	regVBUSVoltageHi = "0x11"
	regVBUSCurrentLo = "0x12"
	regVBUSCurrentHi = "0x13"
	regVBUSPowerLo   = "0x14"
	regVBUSPowerHi   = "0x15"

	regBatteryVoltageLo = "0x20"
	regBatteryVoltageHi = "0x21"
	regBatteryCurrentLo = "0x22"
	regBatteryCurrentHi = "0x23"
	regBatteryPercentLo = "0x24"
	regBatteryPercentHi = "0x25"
	regRemainCapLo      = "0x26"
	regRemainCapHi      = "0x27"
	regRemainDisLo      = "0x28"
	regRemainDisHi      = "0x29"
	regRemainChgLo      = "0x2A"
	regRemainChgHi      = "0x2B"

	regFullCapLo = "0x2C"
	regFullCapHi = "0x2D"

	regCell1Lo = "0x30"
	regCell1Hi = "0x31"
	regCell2Lo = "0x32"
	regCell2Hi = "0x33"
	regCell3Lo = "0x34"
	regCell3Hi = "0x35"
	regCell4Lo = "0x36"
	regCell4Hi = "0x37"

	regFirmwareVersion = "0x50"

	minVBUSPresentMV   = 5000
	minChargeCurrentMA = 100

	i2cTimeoutSec     = 2
	i2cReadRetries    = 3
	i2cRetryDelay     = 150 * time.Millisecond
	i2cRecoverTimeout = 2
)

type rawSnapshot struct {
	CommState   int
	ChargeState int

	VBUSVoltageMV int
	VBUSCurrentMA int
	VBUSPowerMW   int

	BatteryVoltageMV int
	BatteryCurrentMA int
	BatteryPercent   int
	RemainingMAh     int
	FullCapacityMAh  int
	RemainDisMin     int
	RemainChgMin     int

	Cell1MV int
	Cell2MV int
	Cell3MV int
	Cell4MV int

	FirmwareVersion int
	ReadAt          time.Time
}

func Collect() (Response, error) {
	s, err := readRawSnapshot()
	if err != nil {
		return Response{}, err
	}

	return Response{
		OK:          true,
		CollectedAt: s.ReadAt.Format(time.RFC3339),
		Data: Snapshot{
			ModeText:        s.StateText(),
			PowerSourceText: s.PowerSourceText(),
			ChargeText:      s.ChargeDetailsText(),

			VBUSVoltageV: round3(float64(s.VBUSVoltageMV) / 1000),
			VBUSCurrentA: round3(float64(s.VBUSCurrentMA) / 1000),
			VBUSPowerW:   round3(float64(s.VBUSPowerMW) / 1000),

			BatteryVoltageV: round3(float64(s.BatteryVoltageMV) / 1000),
			BatteryCurrentA: round3(float64(s.BatteryCurrentMA) / 1000),
			BatteryPercent:  s.BatteryPercent,
			RemainingMAh:    s.RemainingMAh,
			FullCapacityMAh: s.FullCapacityMAh,

			Cell1MV: s.Cell1MV,
			Cell2MV: s.Cell2MV,
			Cell3MV: s.Cell3MV,
			Cell4MV: s.Cell4MV,

			CellDeltaMV:   s.CellDeltaMV(),
			CellDeltaText: s.CellDeltaText(),

			TimeToChargeMin:     s.RemainChgMin,
			TimeToChargeText:    s.TimeToChargeText(),
			TimeToDischargeMin:  s.RemainDisMin,
			TimeToDischargeText: s.TimeToDischargeText(),
			EtaText:             s.EtaText(),

			CommText:     s.CommText(),
			FirmwareText: s.FirmwareText(),
		},
	}, nil
}

func readRawSnapshot() (*rawSnapshot, error) {
	s := &rawSnapshot{}
	var err error

	if s.CommState, err = readReg8(regCommState); err != nil {
		return nil, fmt.Errorf("read COMM state: %w", err)
	}
	if s.ChargeState, err = readReg8(regChargeState); err != nil {
		return nil, fmt.Errorf("read CHARGE state: %w", err)
	}

	if s.VBUSVoltageMV, err = readU16LE(regVBUSVoltageLo, regVBUSVoltageHi); err != nil {
		return nil, fmt.Errorf("read VBUS voltage: %w", err)
	}
	if s.VBUSCurrentMA, err = readU16LE(regVBUSCurrentLo, regVBUSCurrentHi); err != nil {
		return nil, fmt.Errorf("read VBUS current: %w", err)
	}
	if s.VBUSPowerMW, err = readU16LE(regVBUSPowerLo, regVBUSPowerHi); err != nil {
		return nil, fmt.Errorf("read VBUS power: %w", err)
	}

	if s.BatteryVoltageMV, err = readU16LE(regBatteryVoltageLo, regBatteryVoltageHi); err != nil {
		return nil, fmt.Errorf("read battery voltage: %w", err)
	}
	if s.BatteryCurrentMA, err = readS16LE(regBatteryCurrentLo, regBatteryCurrentHi); err != nil {
		return nil, fmt.Errorf("read battery current: %w", err)
	}
	if s.BatteryPercent, err = readU16LE(regBatteryPercentLo, regBatteryPercentHi); err != nil {
		return nil, fmt.Errorf("read battery percent: %w", err)
	}
	if s.RemainingMAh, err = readU16LE(regRemainCapLo, regRemainCapHi); err != nil {
		return nil, fmt.Errorf("read remaining mAh: %w", err)
	}
	if s.RemainDisMin, err = readU16LE(regRemainDisLo, regRemainDisHi); err != nil {
		return nil, fmt.Errorf("read remain discharge min: %w", err)
	}
	if s.RemainChgMin, err = readU16LE(regRemainChgLo, regRemainChgHi); err != nil {
		return nil, fmt.Errorf("read remain charge min: %w", err)
	}
	if s.FullCapacityMAh, err = readU16LE(regFullCapLo, regFullCapHi); err != nil {
		s.FullCapacityMAh = 0
	}

	if s.Cell1MV, err = readU16LE(regCell1Lo, regCell1Hi); err != nil {
		return nil, fmt.Errorf("read cell1: %w", err)
	}
	if s.Cell2MV, err = readU16LE(regCell2Lo, regCell2Hi); err != nil {
		return nil, fmt.Errorf("read cell2: %w", err)
	}
	if s.Cell3MV, err = readU16LE(regCell3Lo, regCell3Hi); err != nil {
		return nil, fmt.Errorf("read cell3: %w", err)
	}
	if s.Cell4MV, err = readU16LE(regCell4Lo, regCell4Hi); err != nil {
		return nil, fmt.Errorf("read cell4: %w", err)
	}

	if s.FirmwareVersion, err = readReg8(regFirmwareVersion); err != nil {
		s.FirmwareVersion = -1
	}

	s.ReadAt = time.Now()

	return s, nil
}

func (s *rawSnapshot) VBUSPresent() bool {
	return s.VBUSVoltageMV >= minVBUSPresentMV
}

func (s *rawSnapshot) Charging() bool {
	return s.VBUSPresent() && s.BatteryCurrentMA > minChargeCurrentMA
}

func (s *rawSnapshot) Discharging() bool {
	return s.BatteryCurrentMA < 0
}

func (s *rawSnapshot) IdleCharging() bool {
	return s.VBUSPresent() && s.BatteryCurrentMA > 0 && s.BatteryCurrentMA <= minChargeCurrentMA
}

func (s *rawSnapshot) StateText() string {
	switch {
	case s.Charging():
		return "заряджається"
	case s.Discharging():
		return "розряджається"
	case s.IdleCharging():
		return "підтримка заряду"
	case s.VBUSPresent():
		return "підключено до живлення"
	default:
		return "стан невідомий"
	}
}

func (s *rawSnapshot) PowerSourceText() string {
	switch {
	case s.VBUSPresent() && s.Charging():
		return "зовнішнє живлення + зарядка"
	case s.VBUSPresent():
		return "зовнішнє живлення"
	case s.Discharging():
		return "живлення від батареї"
	default:
		return "невідомо"
	}
}

func (s *rawSnapshot) ChargePhase() string {
	phase := (s.ChargeState >> 4) & 0x07

	switch phase {
	case 0:
		return "очікування"
	case 1:
		return "попередній заряд"
	case 2:
		return "постійний струм"
	case 3:
		return "постійна напруга"
	case 4:
		return "заряд завершено"
	case 5:
		return "очікує зарядки"
	case 6:
		return "таймаут зарядки"
	default:
		return "невідомо"
	}
}

func (s *rawSnapshot) IsFastCharging() bool {
	return s.ChargeState&0x80 != 0
}

func (s *rawSnapshot) ChargeDetailsText() string {
	if !s.VBUSPresent() {
		return "—"
	}

	if s.IdleCharging() {
		return "підтримка заряду"
	}

	if !s.Charging() {
		return "заряд не виконується"
	}

	phase := s.ChargePhase()
	if s.IsFastCharging() {
		return phase + " (швидка)"
	}

	return phase
}

func (s *rawSnapshot) CellDeltaMV() int {
	minV := s.Cell1MV
	maxV := s.Cell1MV

	for _, v := range []int{s.Cell2MV, s.Cell3MV, s.Cell4MV} {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}

	return maxV - minV
}

func (s *rawSnapshot) CellDeltaText() string {
	d := s.CellDeltaMV()

	switch {
	case d < 50:
		return fmt.Sprintf("%d mV (ідеально)", d)
	case d < 100:
		return fmt.Sprintf("%d mV (норма)", d)
	case d < 200:
		return fmt.Sprintf("%d mV (є розбаланс)", d)
	default:
		return fmt.Sprintf("%d mV (⚠️ проблема)", d)
	}
}

func (s *rawSnapshot) formatMinutes(mins int) string {
	if mins <= 0 || mins >= 65535 {
		return "—"
	}

	h := mins / 60
	m := mins % 60

	if h > 0 && m > 0 {
		return fmt.Sprintf("%dг %dхв", h, m)
	}

	if h > 0 {
		return fmt.Sprintf("%dг", h)
	}

	return fmt.Sprintf("%dхв", m)
}

func (s *rawSnapshot) TimeToChargeText() string {
	if !s.VBUSPresent() || !s.Charging() {
		return "—"
	}

	return s.formatMinutes(s.RemainChgMin)
}

func (s *rawSnapshot) TimeToDischargeText() string {
	if s.VBUSPresent() && !s.Discharging() {
		return "—"
	}

	if !s.Discharging() {
		return "—"
	}

	return s.formatMinutes(s.RemainDisMin)
}

func (s *rawSnapshot) EtaText() string {
	switch {
	case s.Charging():
		text := s.TimeToChargeText()
		if text == "—" {
			return "До повного заряду: —"
		}
		return "До повного заряду: " + text

	case s.Discharging():
		text := s.TimeToDischargeText()
		if text == "—" {
			return "Час роботи: —"
		}
		return "Час роботи: " + text

	case s.IdleCharging():
		return "Підтримка заряду"

	case s.VBUSPresent():
		return "Живлення підключено"

	default:
		return "ETA недоступний"
	}
}

func (s *rawSnapshot) BQ4050OK() bool {
	return s.CommState&0x01 == 0
}

func (s *rawSnapshot) IP2368OK() bool {
	return s.CommState&0x02 == 0
}

func (s *rawSnapshot) CommText() string {
	bq := "активний"
	if !s.BQ4050OK() {
		bq = "не активний"
	}

	ip := "активний"
	if !s.IP2368OK() {
		if !s.VBUSPresent() {
			ip = "вимкнений"
		} else {
			ip = "не активний"
		}
	}

	return fmt.Sprintf("BQ4050: %s, IP2368: %s", bq, ip)
}

func (s *rawSnapshot) FirmwareText() string {
	if s.FirmwareVersion < 0 {
		return "н/д"
	}
	return fmt.Sprintf("0x%X", s.FirmwareVersion)
}

func readReg8(reg string) (int, error) {
	var lastErr error

	for attempt := 1; attempt <= i2cReadRetries; attempt++ {
		n, err := readReg8Once(reg)
		if err == nil {
			return n, nil
		}

		lastErr = err

		if attempt == 1 {
			recoverI2CBus()
		}

		if attempt < i2cReadRetries {
			time.Sleep(i2cRetryDelay)
		}
	}

	return 0, fmt.Errorf("read reg %s failed after %d attempts: %w", reg, i2cReadRetries, lastErr)
}

func readReg8Once(reg string) (int, error) {
	out, err := runCommand(
		i2cTimeoutSec,
		"i2cget",
		"-y",
		strconv.Itoa(upsI2CBus),
		upsI2CAddr,
		reg,
	)
	if err != nil {
		return 0, err
	}

	n, err := strconv.ParseInt(strings.TrimPrefix(strings.TrimSpace(out), "0x"), 16, 32)
	if err != nil {
		return 0, fmt.Errorf("parse %q: %w", out, err)
	}

	return int(n), nil
}

func recoverI2CBus() {
	_, _ = runCommand(
		i2cRecoverTimeout,
		"i2cdetect",
		"-y",
		strconv.Itoa(upsI2CBus),
	)
}

func readU16LE(lo, hi string) (int, error) {
	l, err := readReg8(lo)
	if err != nil {
		return 0, err
	}
	h, err := readReg8(hi)
	if err != nil {
		return 0, err
	}
	return (h << 8) | l, nil
}

func readS16LE(lo, hi string) (int, error) {
	v, err := readU16LE(lo, hi)
	if err != nil {
		return 0, err
	}
	if v >= 32768 {
		return v - 65536, nil
	}
	return v, nil
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

func round3(v float64) float64 {
	return float64(int(v*1000+0.5)) / 1000
}
