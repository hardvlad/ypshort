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
	mu     sync.RWMutex
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

func (s *Storage) Set(key, value string, userID int) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Get(key); exists {
		return "", false, fmt.Errorf("%w: %s", repository.ErrorKeyExists, key)
	}

	_, err := s.DBConn.ExecContext(context.Background(), "INSERT INTO saved_links (code, url, user_id) VALUES ($1, $2, $3)", key, value, userID)
	if err != nil {
		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			if code, exists := s.GetCode(value); exists {
				return code, true, nil
			}
		}
		return "", false, err
	}
	return key, false, nil
}

func (s *Storage) GetUserData(userID int) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.DBConn.QueryContext(context.Background(), "SELECT code, url from saved_links where user_id = $1", userID)
	if err != nil {
		s.logger.Debugw(err.Error(), "event", "получение данных пользователя", "user_id", userID)
		return nil, err
	}
	defer rows.Close()

	userData := make(map[string]string)
	for rows.Next() {
		var code, url string
		if err := rows.Scan(&code, &url); err != nil {
			return nil, err
		}
		userData[code] = url
	}
	return userData, nil
}
