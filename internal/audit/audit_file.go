package audit

import "os"

type AuditorFile struct {
	ID       string
	FilePath string
}

func InitAuditFile(path string) *AuditorFile {
	return &AuditorFile{
		ID:       "auditFile",
		FilePath: path,
	}
}

func (s *AuditorFile) Update(data AuditorEvent) {
	appendStringToFile(s.FilePath, data.GetJSON())
}

func (s *AuditorFile) getID() string {
	return s.ID
}

func appendStringToFile(filePath string, data string) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	if _, err := f.WriteString(data); err != nil {
		return
	}
}
