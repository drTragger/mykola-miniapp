package metrics

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	historyDBPath         = "storage/metrics-history.db"
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

	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			_ = db.Close()
			return fmt.Errorf("apply pragma %q: %w", p, err)
		}
	}

	schema := `
	CREATE TABLE IF NOT EXISTS metrics_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sample_bucket INTEGER NOT NULL UNIQUE,
		collected_at INTEGER NOT NULL,
		cpu_usage_percent REAL NOT NULL DEFAULT 0,
		cpu_temp_celsius REAL NOT NULL DEFAULT 0,
		ram_usage_percent REAL NOT NULL DEFAULT 0,
		rx_speed_bps REAL NOT NULL DEFAULT 0,
		tx_speed_bps REAL NOT NULL DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_metrics_history_collected_at
	ON metrics_history (collected_at);
	`

	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return fmt.Errorf("create schema: %w", err)
	}

	historyDB = db
	return nil
}

func storeHistorySnapshot(r Response, collectedAt time.Time) error {
	if err := initHistoryDB(); err != nil {
		return err
	}

	bucket := collectedAt.Truncate(historySampleInterval).Unix()

	result, err := historyDB.Exec(`
		INSERT OR IGNORE INTO metrics_history (
			sample_bucket, collected_at,
			cpu_usage_percent, cpu_temp_celsius, ram_usage_percent,
			rx_speed_bps, tx_speed_bps
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		bucket,
		collectedAt.Unix(),
		r.Overview.CPUUsagePercent,
		r.Overview.CPUTemperatureCelsius,
		r.Overview.RAMUsagePercent,
		r.Network.RxSpeedBps,
		r.Network.TxSpeedBps,
	)
	if err != nil {
		return fmt.Errorf("insert metrics history: %w", err)
	}

	rows, err := result.RowsAffected()
	if err == nil && rows > 0 {
		return cleanupOldHistory()
	}

	return nil
}

func cleanupOldHistory() error {
	cutoff := time.Now().Add(-time.Duration(historyRetentionDays) * 24 * time.Hour).Unix()
	_, err := historyDB.Exec(`DELETE FROM metrics_history WHERE collected_at < ?`, cutoff)
	if err != nil {
		return fmt.Errorf("cleanup metrics history: %w", err)
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
		SELECT collected_at, cpu_usage_percent, cpu_temp_celsius, ram_usage_percent,
		       rx_speed_bps, tx_speed_bps
		FROM (
			SELECT collected_at, cpu_usage_percent, cpu_temp_celsius, ram_usage_percent,
			       rx_speed_bps, tx_speed_bps
			FROM metrics_history
			ORDER BY collected_at DESC
			LIMIT ?
		)
		ORDER BY collected_at ASC
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query metrics history: %w", err)
	}
	defer rows.Close()

	points := make([]HistoryPoint, 0, limit)

	for rows.Next() {
		var ts int64
		var p HistoryPoint

		if err := rows.Scan(&ts, &p.CPUUsagePercent, &p.CPUTempCelsius, &p.RAMUsagePercent, &p.RxSpeedBps, &p.TxSpeedBps); err != nil {
			return nil, fmt.Errorf("scan metrics history: %w", err)
		}

		p.Timestamp = ts
		p.Time = time.Unix(ts, 0).Format("15:04")
		points = append(points, p)
	}

	return points, rows.Err()
}
