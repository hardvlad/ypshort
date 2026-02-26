package audit

import (
	"os"
	"sync"
)

// AuditorFile is an audit implementation that writes events to a file.
type AuditorFile struct {
	ID       string
	FilePath string
	file     *os.File
	mu       sync.Mutex
}

// InitAuditFile creates a new AuditorFile object
func InitAuditFile(path string) (*AuditorFile, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &AuditorFile{
		ID:       "auditFile",
		FilePath: path,
		file:     f,
	}, nil
}

// Update appends the event to the file.
func (s *AuditorFile) Update(data AuditorEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file != nil {
		s.file.WriteString(data.GetJSON())
	}
}

func (s *AuditorFile) getID() string {
	return s.ID
}

// Close closes the file. Should be called when shutting down.
func (s *AuditorFile) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file != nil {
		err := s.file.Close()
		s.file = nil
		return err
	}
	return nil
}
