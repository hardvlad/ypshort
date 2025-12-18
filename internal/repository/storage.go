package repository

import (
	"errors"
	"sync"
)

type Storage struct {
	kvStorage map[string]string
	mu        sync.Mutex
}

func NewStorage() *Storage {
	return &Storage{
		kvStorage: make(map[string]string),
	}
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
		return ErrorKeyExists
	}
	s.kvStorage[key] = value
	return nil
}
