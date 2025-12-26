package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
)

type StorageInterface interface {
	Get(key string) (string, bool, bool)
	Set(key, value string, userID int) (string, bool, error)
	GetUserData(userID int) (map[string]string, error)
	DeleteURLs(codes []string, userID int) error
}

type Storage struct {
	kvStorage   map[string]string
	mu          sync.Mutex
	fileName    string
	sugarLogger *zap.SugaredLogger
}

func (s *Storage) DeleteURLs(codes []string, userID int) error {
	for _, code := range codes {
		delete(s.kvStorage, code)
	}
	return nil
}

type JSONParseMap struct {
	Data map[string]string `json:",unknown"`
}

func NewStorage(fileName string, sugarLogger *zap.SugaredLogger) (*Storage, error) {
	if fileName == "" {
		return makeEmptyStorage(fileName, sugarLogger)
	}

	if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
		return makeEmptyStorage(fileName, sugarLogger)
	}

	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, fileName)
	}
	defer file.Close()

	var data JSONParseMap
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, err
	}

	if data.Data != nil {
		return &Storage{
			kvStorage:   data.Data,
			fileName:    fileName,
			sugarLogger: sugarLogger,
		}, nil
	}

	return makeEmptyStorage(fileName, sugarLogger)
}

func makeEmptyStorage(fileName string, sugarLogger *zap.SugaredLogger) (*Storage, error) {
	return &Storage{
		kvStorage:   make(map[string]string),
		fileName:    fileName,
		sugarLogger: sugarLogger,
	}, nil
}

func (s *Storage) Get(key string) (string, bool, bool) {
	value, ok := s.kvStorage[key]
	return value, false, ok
}

var ErrorKeyExists = errors.New("key already exists")

func (s *Storage) Set(key, value string, userID int) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.kvStorage[key]; exists {
		return key, false, fmt.Errorf("%w: %s", ErrorKeyExists, key)
	}
	s.kvStorage[key] = value

	err := s.persistToFile()
	if err != nil && s.sugarLogger != nil {
		s.sugarLogger.Errorw("ошибка записи в базу", "err", err.Error())
	}

	return key, false, nil
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

	data := JSONParseMap{Data: s.kvStorage}
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
