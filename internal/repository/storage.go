package repository

type Storage struct {
	kvStorage map[string]string
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

func (s *Storage) Set(key, value string) {
	s.kvStorage[key] = value
}
