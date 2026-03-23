package ups

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	historyDBPath         = "storage/ups-history.db"
	historySampleInterval = 5 * time.Minute
	historyRetentionDays  = 30
	historyDefaultLimit   = 288
)

var historyDB *sql.DB

func initHistoryDB() error {
	if historyDB != nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(historyDBPath), 0o755); err != nil {
		return fmt.Errorf("create history dir: %w", err)
	}

	db, err := sql.Open("sqlite3", historyDBPath)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	pragmas := []string{
		`PRAGMA journal_mode = WAL;`,
		`PRAGMA synchronous = NORMAL;`,
		`PRAGMA busy_timeout = 5000;`,
		`PRAGMA temp_store = MEMORY;`,
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			_ = db.Close()
			return fmt.Errorf("apply pragma %q: %w", pragma, err)
		}
	}

	schema := `
	CREATE TABLE IF NOT EXISTS ups_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sample_bucket INTEGER NOT NULL UNIQUE,
		collected_at INTEGER NOT NULL,
		battery_percent INTEGER NOT NULL,
		cell_delta_mv INTEGER NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_ups_history_collected_at
	ON ups_history (collected_at);
	`

	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return fmt.Errorf("create schema: %w", err)
	}

	historyDB = db

	return nil
}

func storeHistorySnapshot(s Snapshot, collectedAt time.Time) error {
	if err := initHistoryDB(); err != nil {
		return err
	}

	bucket := collectedAt.Truncate(historySampleInterval).Unix()
	collectedUnix := collectedAt.Unix()

	result, err := historyDB.Exec(`
		INSERT OR IGNORE INTO ups_history (
			sample_bucket,
			collected_at,
			battery_percent,
			cell_delta_mv
		) VALUES (?, ?, ?, ?)
	`,
		bucket,
		collectedUnix,
		s.BatteryPercent,
		s.CellDeltaMV,
	)
	if err != nil {
		return fmt.Errorf("insert history point: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected > 0 {
		if err := cleanupOldHistory(); err != nil {
			return err
		}
	}

	return nil
}

func cleanupOldHistory() error {
	if historyDB == nil {
		return nil
	}

	cutoff := time.Now().Add(-time.Duration(historyRetentionDays) * 24 * time.Hour).Unix()

	_, err := historyDB.Exec(`
		DELETE FROM ups_history
		WHERE collected_at < ?
	`, cutoff)
	if err != nil {
		return fmt.Errorf("cleanup old history: %w", err)
	}

	return nil
}

func InitHistory() error {
	return initHistoryDB()
}

func GetHistory(limit int) ([]HistoryPoint, error) {
	if err := initHistoryDB(); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = historyDefaultLimit
	}

	rows, err := historyDB.Query(`
		SELECT collected_at, battery_percent, cell_delta_mv
		FROM (
			SELECT collected_at, battery_percent, cell_delta_mv
			FROM ups_history
			ORDER BY collected_at DESC
			LIMIT ?
		)
		ORDER BY collected_at ASC
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()

	points := make([]HistoryPoint, 0, limit)

	for rows.Next() {
		var ts int64
		var batteryPercent int
		var cellDeltaMV int

		if err := rows.Scan(&ts, &batteryPercent, &cellDeltaMV); err != nil {
			return nil, fmt.Errorf("scan history point: %w", err)
		}

		t := time.Unix(ts, 0)

		points = append(points, HistoryPoint{
			Timestamp:      ts,
			BatteryPercent: batteryPercent,
			CellDeltaMV:    cellDeltaMV,
			Time:           t.Format("15:04"),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history rows: %w", err)
	}

	return points, nil
}

func StartHistoryCollector() {
	save := func() {
		resp, err := GetSnapshot()
		if err != nil {
			return
		}

		collectedAt, err := time.Parse(time.RFC3339, resp.CollectedAt)
		if err != nil {
			collectedAt = time.Now()
		}

		_ = storeHistorySnapshot(resp.Data, collectedAt)
	}

	save()

	ticker := time.NewTicker(historySampleInterval)
	defer ticker.Stop()

	for range ticker.C {
		save()
	}
}
