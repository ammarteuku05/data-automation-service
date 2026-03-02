package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// conn db postgres
func NewPostgresDB(url string, maxOpenConns, maxIdleConns int) (*sqlx.DB, error) {
	if url == "" {
		return nil, fmt.Errorf("database URL is empty")
	}

	db, err := sqlx.Connect("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	if maxOpenConns > 0 {
		db.SetMaxOpenConns(maxOpenConns)
	}
	if maxIdleConns > 0 {
		db.SetMaxIdleConns(maxIdleConns)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging the database: %w", err)
	}

	return db, nil
}
