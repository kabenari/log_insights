package storage

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/kabenari/log-insight/pkg/models"
	_ "github.com/mattn/go-sqlite3" // Ensure this import is here
)

// This is the global variable. It stays nil if you use ":=" inside InitDB.
var DB *sql.DB

func InitDB() {
	dbFile := "./insights.db"
	absPath, _ := filepath.Abs(dbFile)
	fmt.Printf("Database path: %s\n", absPath)

	var err error
	// CORRECT: Use "=" (assignment), NOT ":=" (declaration)
	DB, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}

	// Verify connection immediately to be sure
	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping DB:", err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS insights (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME,
		level TEXT,
		message TEXT,
		service TEXT,
		status INTEGER,
		analysis TEXT,
		fixed BOOLEAN
	);`

	_, err = DB.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
}

func SaveInsight(result models.AIResult) error {
	// If DB is nil here, it crashes. The fix above prevents this.
	stmt, err := DB.Prepare("INSERT INTO insights(timestamp, level, message, service, status, analysis, fixed) VALUES(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		result.OriginalLog.Timestamp,
		result.OriginalLog.Level,
		result.OriginalLog.Message,
		result.OriginalLog.Service,
		result.OriginalLog.Status,
		result.Analysis,
		result.Fixed,
	)
	return err
}

func GetInsights() []models.AIResult {
	// Simple safety check
	if DB == nil {
		return []models.AIResult{}
	}

	rows, err := DB.Query("SELECT timestamp, level, message, service, status, analysis, fixed FROM insights ORDER BY id DESC LIMIT 50")
	if err != nil {
		log.Println("Error querying insights:", err)
		return []models.AIResult{}
	}
	defer rows.Close()

	var results []models.AIResult
	for rows.Next() {
		var r models.AIResult
		// You might need a temporary variable for timestamp scanning if it fails
		// but standard sql driver usually handles time.Time
		err = rows.Scan(
			&r.OriginalLog.Timestamp,
			&r.OriginalLog.Level,
			&r.OriginalLog.Message,
			&r.OriginalLog.Service,
			&r.OriginalLog.Status,
			&r.Analysis,
			&r.Fixed,
		)
		if err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		results = append(results, r)
	}
	return results
}
