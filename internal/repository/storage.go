package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

type Storage struct {
	kvStorage map[string]string
	mu        sync.Mutex
	fileName  string
}

type JSONParseMap struct {
	Data map[string]string `json:",unknown"`
}

func NewStorage(fileName string) (*Storage, error) {
	if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
		return makeEmptyStorage(fileName)
	}

	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data JSONParseMap
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, err
	}

	if data.Data != nil {
		return &Storage{
			kvStorage: data.Data,
			fileName:  fileName,
		}, nil
	}

	return makeEmptyStorage(fileName)
}

func makeEmptyStorage(fileName string) (*Storage, error) {
	return &Storage{
		kvStorage: make(map[string]string),
		fileName:  fileName,
	}, nil
}

func (s *Storage) Get(key string) (string, bool) {
	value, ok := s.kvStorage[key]
	return value, ok
}

var ErrorKeyExists = errors.New("key already exists")

func (s *Storage) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.kvStorage[key]; exists {
		return fmt.Errorf("%w: %s", ErrorKeyExists, key)
	}
	s.kvStorage[key] = value

	s.persistToFile()

	return nil
}

func (s *Storage) persistToFile() {

	file, err := os.OpenFile(s.fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return
	}
	defer file.Close()

	data := JSONParseMap{Data: s.kvStorage}
	encoder := json.NewEncoder(file)
	_ = encoder.Encode(data)
}
