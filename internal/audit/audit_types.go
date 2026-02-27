package audit

import "encoding/json"

// AuditorEvent contains audit event data
type AuditorEvent struct {
	TS     int64  `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id"`
	URL    string `json:"url"`
}

// GetJSON returns the event data as JSON string
func (s AuditorEvent) GetJSON() string {
	data, _ := json.Marshal(s)
	return string(data)
}
