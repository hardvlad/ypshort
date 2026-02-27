package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// StorageInterface defines methods for managing storage of key-value pairs and associated user data.
type StorageInterface interface {
	Get(key string) (string, bool, bool)
	Set(key, value string, userID int) (string, bool, error)
	GetUserData(userID int) (map[string]string, error)
	DeleteURLs(codes []string, userID int) error
	Close() error
}

// Storage represents a thread-safe key-value storage with persistent file support and context-based lifecycle management.
type Storage struct {
	kvStorage   map[string]string
	mu          sync.Mutex
	fileName    string
	sugarLogger *zap.SugaredLogger

	hasChanges atomic.Bool
	persistCh  chan struct{}
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

type jsonParseMap struct {
	Data map[string]string `json:",unknown"`
}

// ErrorKeyExists indicates that an attempt was made to insert a duplicate key into the storage.
var ErrorKeyExists = errors.New("key already exists")

// NewStorage creates a new Storage instance with the specified file name and logger, managing persistence and lifecycle.
func NewStorage(fileName string, sugarLogger *zap.SugaredLogger) (*Storage, error) {

	var s *Storage
	var err error

	if fileName == "" {
		s, err = makeEmptyStorage(fileName, sugarLogger)
	} else {
		if _, err1 := os.Stat(fileName); errors.Is(err1, os.ErrNotExist) {
			s, err = makeEmptyStorage(fileName, sugarLogger)
		} else {
			file, err1 := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
			if err1 != nil {
				return nil, fmt.Errorf("%w: %s", err1, fileName)
			}
			defer file.Close()

			var data jsonParseMap
			if err1 := json.NewDecoder(file).Decode(&data); err1 != nil {
				return nil, err1
			}

			if data.Data != nil {
				ctx, cancel := context.WithCancel(context.Background())
				s, err = &Storage{
					kvStorage:   data.Data,
					fileName:    fileName,
					sugarLogger: sugarLogger,
					persistCh:   make(chan struct{}, 1),
					ctx:         ctx,
					cancel:      cancel,
				}, nil
			} else {
				s, err = makeEmptyStorage(fileName, sugarLogger)
			}
		}
	}

	if err == nil && s != nil {
		s.wg.Add(1)
		go s.periodicPersistLoop(10 * time.Second)
	}

	return s, err
}

// Get retrieves the value associated with the provided key and returns it along with two boolean flags.
// The first boolean flag indicates if the value is invalid, while the second indicates if the key exists.
func (s *Storage) Get(key string) (string, bool, bool) {
	value, ok := s.kvStorage[key]
	return value, false, ok
}

// Set attempts to store a key-value pair in the storage if the key does not already exist, ensuring thread-safety.
func (s *Storage) Set(key, value string, userID int) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.kvStorage[key]; exists {
		return key, false, fmt.Errorf("%w: %s", ErrorKeyExists, key)
	}
	s.kvStorage[key] = value

	// Mark that the data has been changed
	s.hasChanges.Store(true)

	// Send the change event to save data
	select {
	case s.persistCh <- struct{}{}:
	default:
	}

	return key, false, nil
}

// DeleteURLs removes the specified list of URL codes from the storage associated with the given user ID.
func (s *Storage) DeleteURLs(codes []string, userID int) error {
	for _, code := range codes {
		delete(s.kvStorage, code)
	}
	return nil
}

func (s *Storage) periodicPersistLoop(d time.Duration) {
	defer s.wg.Done()

	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			if s.hasChanges.Load() {
				_ = s.persistToFile()
			}
			return

		case <-ticker.C:
			s.tryPersistLocked()

		case <-s.persistCh:
			//time.Sleep(20 * time.Millisecond)
			s.tryPersistLocked()
		}
	}
}

func (s *Storage) tryPersistLocked() {
	if s.hasChanges.CompareAndSwap(true, false) {
		if err := s.persistToFile(); err != nil {
			s.sugarLogger.Errorw("ошибка записи в базу", "err", err.Error())
		}
	}
}

func makeEmptyStorage(fileName string, sugarLogger *zap.SugaredLogger) (*Storage, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &Storage{
		kvStorage:   make(map[string]string, 1000),
		fileName:    fileName,
		sugarLogger: sugarLogger,
		persistCh:   make(chan struct{}, 1),
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

func (s *Storage) persistToFile() error {
	if s.fileName == "" {
		return errors.New("файл базы не задан")
	}

	file, err := os.OpenFile(s.fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return errors.New("не могу открыть файл для записи: " + s.fileName)
	}
	defer file.Close()

	data := jsonParseMap{Data: s.kvStorage}
	encoder := json.NewEncoder(file)
	err = encoder.Encode(data)
	if err != nil {
		return fmt.Errorf("ошибка сериализации в базу %w: %s", err, s.fileName)
	}

	return nil
}

func (s *Storage) GetUserData(userID int) (map[string]string, error) {
	return s.kvStorage, nil
}

func (s *Storage) Close() error {
	s.cancel()
	s.wg.Wait()
	return nil
}
