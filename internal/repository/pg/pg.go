package pg

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/hardvlad/ypshort/internal/repository"
	"go.uber.org/zap"
)

type Storage struct {
	DbConn *sql.DB
	mu     sync.Mutex
	logger *zap.SugaredLogger
}

func NewPGStorage(dbConn *sql.DB, logger *zap.SugaredLogger) Storage {
	return Storage{DbConn: dbConn, logger: logger}
}

func (s Storage) Get(key string) (string, bool) {
	row := s.DbConn.QueryRowContext(context.Background(), "SELECT url from saved_links where code = $1 limit 1", key)

	var savedUrl string
	err := row.Scan(&savedUrl)
	if err != nil {
		s.logger.Debugw(err.Error(), "event", "get from DB", key)
		return "", false
	}
	return savedUrl, true
}

func (s Storage) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Get(key); exists {
		return fmt.Errorf("%w: %s", repository.ErrorKeyExists, key)
	}

	_, err := s.DbConn.ExecContext(context.Background(), "INSERT INTO saved_links (code, url) VALUES ($1, $2)", key, value)
	if err != nil {
		return err
	}
	return nil
}
