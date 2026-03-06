package service

import (
	"context"
	"errors"
	"math/rand"
	"net/url"
	"time"

	"github.com/hardvlad/ypshort/internal/audit"
	"github.com/hardvlad/ypshort/internal/config"
	"github.com/hardvlad/ypshort/internal/repository"
	"go.uber.org/zap"
)

type URLData struct {
	ID          string    `json:"id"`
	OriginalURL string    `json:"original_url"`
	CreatorID   string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
}

type Storage interface {
	Save(ctx context.Context, data *URLData) error
	GetByID(ctx context.Context, id string) (*URLData, error)
	GetByCreator(ctx context.Context, creatorID string) ([]*URLData, error)
}

type AuthService interface {
	GetUserID(ctx context.Context) (string, bool)
}

// ShortenerService реализует бизнес-логику
type ShortenerService struct {
	Conf     *config.Config
	Store    repository.StorageInterface
	Logger   *zap.SugaredLogger
	Observer *audit.Event
}

func NewShortenerService(ctx context.Context, conf *config.Config, store repository.StorageInterface, sugarLogger *zap.SugaredLogger, observer *audit.Event) *ShortenerService {
	return &ShortenerService{
		Conf:     conf,
		Store:    store,
		Logger:   sugarLogger,
		Observer: observer,
	}
}

// Shorten создаёт новую короткую ссылку
func (s *ShortenerService) Shorten(urlToShort string, userID int) (success bool, code string, urlExisted bool, err error) {
	success, shortLink, urlAlreadyExisted, err := s.getShortCode(urlToShort, 5, userID)
	fullURL := ""
	if success {
		fullURL, err = url.JoinPath(s.Conf.ServerAddress, shortLink)
		if err != nil {
			return false, "", false, err
		}
	}
	return success, fullURL, urlAlreadyExisted, err
}

func (s *ShortenerService) getShortCode(url string, maxAttempts int, userID int) (success bool, code string, urlExisted bool, err error) {
	success = false
	var shortLink string
	var urlAlreadyExisted bool

	for i := 0; i < maxAttempts; i++ {
		shortLink = s.generateRandomString()
		code, urlExisted, err := s.Store.Set(shortLink, url, userID)
		if err != nil {
			if errors.Is(err, repository.ErrorKeyExists) {
				continue
			} else {
				return false, "", false, err
			}
		} else {
			success = true
			shortLink = code
			urlAlreadyExisted = urlExisted
			break
		}
	}
	return success, shortLink, urlAlreadyExisted, err
}

func (s *ShortenerService) generateRandomString() string {
	b := make([]byte, s.Conf.ShortLinkLength)
	for i := 0; i < s.Conf.ShortLinkLength; i++ {
		b[i] = s.Conf.Charset[rand.Intn(len(s.Conf.Charset))]
	}
	return string(b[:])
}

// Expand возвращает оригинальный URL по ID
func (s *ShortenerService) Expand(id string) (string, bool, bool) {
	return s.Store.Get(id)
}

// ListUserURLs возвращает список ссылок пользователя
func (s *ShortenerService) ListUserURLs(ctx context.Context, userID int) (map[string]string, error) {
	return s.Store.GetUserData(userID)
}

var (
	ErrNotFound     = &APIError{code: 404, message: "short URL not found"}
	ErrBadRequest   = &APIError{code: 400, message: "invalid request"}
	ErrUnauthorized = &APIError{code: 401, message: "unauthorized"}
)

type APIError struct {
	code    int
	message string
}

func (e *APIError) Error() string { return e.message }
func (e *APIError) Code() int     { return e.code }
