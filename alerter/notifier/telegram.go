package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Telegram struct {
	botToken string
	chatID   string
	client   *http.Client
}

func (t *Telegram) getClient() *http.Client {
	if t.client == nil {
		t.client = &http.Client{Timeout: 10 * time.Second}
	}
	return t.client
}

func (t *Telegram) Send(subject, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
	payload := map[string]string{
		"chat_id":    t.chatID,
		"text":       fmt.Sprintf("*%s*\n\n%s", subject, message),
		"parse_mode": "Markdown",
	}
	body, _ := json.Marshal(payload)
	resp, err := t.getClient().Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram returned %d", resp.StatusCode)
	}
	return nil
}
