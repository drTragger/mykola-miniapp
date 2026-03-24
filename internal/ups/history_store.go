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
		cell_delta_mv INTEGER NOT NULL,
		cell1_mv INTEGER NOT NULL DEFAULT 0,
		cell2_mv INTEGER NOT NULL DEFAULT 0,
		cell3_mv INTEGER NOT NULL DEFAULT 0,
		cell4_mv INTEGER NOT NULL DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_ups_history_collected_at
	ON ups_history (collected_at);
	`

	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return fmt.Errorf("create schema: %w", err)
	}

	migrations := []string{
		`ALTER TABLE ups_history ADD COLUMN cell1_mv INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE ups_history ADD COLUMN cell2_mv INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE ups_history ADD COLUMN cell3_mv INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE ups_history ADD COLUMN cell4_mv INTEGER NOT NULL DEFAULT 0;`,
	}

	for _, migration := range migrations {
		_, _ = db.Exec(migration)
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
			cell_delta_mv,
			cell1_mv,
			cell2_mv,
			cell3_mv,
			cell4_mv
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		bucket,
		collectedUnix,
		s.BatteryPercent,
		s.CellDeltaMV,
		s.Cell1MV,
		s.Cell2MV,
		s.Cell3MV,
		s.Cell4MV,
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
		SELECT collected_at, battery_percent, cell_delta_mv, cell1_mv, cell2_mv, cell3_mv, cell4_mv
		FROM (
			SELECT collected_at, battery_percent, cell_delta_mv, cell1_mv, cell2_mv, cell3_mv, cell4_mv
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
		var cell1MV int
		var cell2MV int
		var cell3MV int
		var cell4MV int

		if err := rows.Scan(
			&ts,
			&batteryPercent,
			&cellDeltaMV,
			&cell1MV,
			&cell2MV,
			&cell3MV,
			&cell4MV,
		); err != nil {
			return nil, fmt.Errorf("scan history point: %w", err)
		}

		t := time.Unix(ts, 0)

		points = append(points, HistoryPoint{
			Timestamp:      ts,
			BatteryPercent: batteryPercent,
			CellDeltaMV:    cellDeltaMV,
			Cell1MV:        cell1MV,
			Cell2MV:        cell2MV,
			Cell3MV:        cell3MV,
			Cell4MV:        cell4MV,
			Time:           t.Format("15:04"),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history rows: %w", err)
	}

	return points, nil
}
