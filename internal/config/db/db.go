// Package db creates connection to a database
package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Config contains database connection parameters
type Config struct {
	Dsn string
}

// NewConfig creates a new Config object
func NewConfig(dsn string) *Config {
	return &Config{
		Dsn: dsn,
	}
}

// InitDB creates a new database connection
func (c *Config) InitDB() (*sql.DB, error) {

	if c.Dsn == "" {
		return nil, fmt.Errorf("dsn не задано")
	}

	db, err := sql.Open("pgx", c.Dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка соединения с базой данных: %w", err)
	}

	// Verify the connection is alive
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка пинга базы данных: %w", err)
	}

	return db, nil
}
