package audit

import (
	"net/http"
	"strings"

	"github.com/hardvlad/ypshort/internal/retry"
)

// AuditorURL is an auditor that sends events to the specified URL.
type AuditorURL struct {
	ID  string
	URL string
}

// InitAuditURL creates a new AuditorURL object
func InitAuditURL(url string) *AuditorURL {
	return &AuditorURL{
		ID:  "auditURL",
		URL: url,
	}
}

// Update appends the event to the URL.
func (s *AuditorURL) Update(data AuditorEvent) {
	go postDataToURL(s.URL, data.GetJSON())
}

func (s *AuditorURL) getID() string {
	return s.ID
}

func postDataToURL(URL string, data string) {
	post, err := retry.Retry(3, 2,
		func() (*http.Response, error) {
			return http.Post(URL, "application/json; charset=utf-8", strings.NewReader(data))
		})

	if err != nil {
		return
	}
	post.Body.Close()
}
