package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Slack struct {
	webhookURL string
	client     *http.Client
}

func (s *Slack) client_() *http.Client {
	if s.client == nil {
		s.client = &http.Client{Timeout: 10 * time.Second}
	}
	return s.client
}

func (s *Slack) Send(subject, message string) error {
	payload := map[string]interface{}{
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]string{"type": "plain_text", "text": subject},
			},
			{
				"type": "section",
				"text": map[string]string{"type": "mrkdwn", "text": message},
			},
		},
	}
	body, _ := json.Marshal(payload)
	resp, err := s.client_().Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("slack returned %d", resp.StatusCode)
	}
	return nil
}
