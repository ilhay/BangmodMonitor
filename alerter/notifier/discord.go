package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Discord struct {
	webhookURL string
	client     *http.Client
}

func (d *Discord) getClient() *http.Client {
	if d.client == nil {
		d.client = &http.Client{Timeout: 10 * time.Second}
	}
	return d.client
}

func (d *Discord) Send(subject, message string) error {
	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       subject,
				"description": message,
				"color":       0xFF4444,
			},
		},
	}
	body, _ := json.Marshal(payload)
	resp, err := d.getClient().Post(d.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("discord returned %d", resp.StatusCode)
	}
	return nil
}
