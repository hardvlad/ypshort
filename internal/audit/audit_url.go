package audit

import (
	"net/http"
	"strings"
)

type AuditorURL struct {
	Id  string
	URL string
}

func InitAuditURL(url string) *AuditorURL {
	return &AuditorURL{
		Id:  "auditURL",
		URL: url,
	}
}

func (s *AuditorURL) Update(data AuditorEvent) {
	go postDataToURL(s.URL, data.GetJSON())
}

func (s *AuditorURL) getID() string {
	return s.Id
}

func postDataToURL(URL string, data string) {
	post, err := http.Post(URL, "application/json; charset=utf-8", strings.NewReader(data))
	if err != nil {
		return
	}
	defer post.Body.Close()
}
