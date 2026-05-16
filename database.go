package rinhabackend2026

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DetectedFraud struct {
	ID          int
	Timestamp   time.Time
	Vector      []float64
	FraudScore  float64
	Approved    bool
	PayloadHash string
}

func InitDB() *sql.DB {

	db, err := sql.Open("sqlite3", "./frauds.db?cache=shared&_journal_mode=WAL")
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS detected_frauds (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        vector TEXT NOT NULL,
        fraud_score REAL NOT NULL,
        approved BOOLEAN NOT NULL,
        payload_hash TEXT UNIQUE
	);

	CREATE INDEX IF NOT EXISTS idx_timestamp ON detected_frauds(timestamp);
	CREATE INDEX IF NOT EXISTS idx_score ON detected_frauds(timestamp);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatal(err)
	}

	return db
}

func InsertDetectedFraud(db *sql.DB, vector []float64, fraudScore float64, approved bool, payloadHash string) error {
	vectorJSON, err := json.Marshal(vector)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO detected_frauds (vector, fraud_score, approved, payload_hash)
		VALUES (?, ?, ?, ?)
	`, string(vectorJSON), fraudScore, approved, payloadHash)

	return err
}
