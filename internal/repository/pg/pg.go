package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/hardvlad/ypshort/internal/repository"
	"go.uber.org/zap"
)

type Storage struct {
	DBConn *sql.DB
	mu     sync.Mutex
	logger *zap.SugaredLogger
}

func NewPGStorage(dbConn *sql.DB, logger *zap.SugaredLogger) *Storage {
	return &Storage{DBConn: dbConn, logger: logger}
}

func (s *Storage) Get(key string) (string, bool) {
	row := s.DBConn.QueryRowContext(context.Background(), "SELECT url from saved_links where code = $1 limit 1", key)

	var savedURL string
	err := row.Scan(&savedURL)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			s.logger.Debugw(err.Error(), "event", "get from DB", "code", key)
		}
		return "", false
	}
	return savedURL, true
}

func (s *Storage) GetCode(url string) (string, bool) {
	row := s.DBConn.QueryRowContext(context.Background(), "SELECT code from saved_links where url = $1 limit 1", url)

	var savedCode string
	err := row.Scan(&savedCode)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			s.logger.Debugw(err.Error(), "event", "get code from DB", "url", url)
		}
		return "", false
	}
	return savedCode, true
}

func (s *Storage) Set(key, value string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Get(key); exists {
		return "", false, fmt.Errorf("%w: %s", repository.ErrorKeyExists, key)
	}

	_, err := s.DBConn.ExecContext(context.Background(), "INSERT INTO saved_links (code, url) VALUES ($1, $2)", key, value)
	if err != nil {
		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			if code, exists := s.GetCode(value); exists {
				return code, true, nil
			}
		} else {
			fmt.Println("Error not PG %w", err)
		}
		return "", false, err
	}
	return key, false, nil
}
