package audit

import "encoding/json"

type AuditorEvent struct {
	Ts     int64  `json:"ts"`
	Action string `json:"action"`
	UserId string `json:"user_id"`
	Url    string `json:"url"`
}

func (s AuditorEvent) GetJSON() string {
	data, _ := json.Marshal(s)
	return string(data)
}
