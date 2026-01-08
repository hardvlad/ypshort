package audit

import "encoding/json"

type AuditorEvent struct {
	TS     int64  `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id"`
	URL    string `json:"url"`
}

func (s AuditorEvent) GetJSON() string {
	data, _ := json.Marshal(s)
	return string(data)
}
